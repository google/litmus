# Copyright 2024 Google, LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Guards for outbound test requests made by the worker.

Test templates let users point the worker at arbitrary HTTP endpoints, which is
the purpose of the tool. The one destination that must never be reachable is
the Google Cloud metadata server: a request to it would return the worker's
own service-account credentials, which would then be stored as the test
"response" and become readable through the Litmus API (SSRF -> credential
theft). This module blocks that destination, including via redirects, and
refuses the ``Metadata-Flavor`` header the metadata server requires.
"""

import ipaddress
import socket
from urllib.parse import urlsplit

import requests

# Hostnames under which the metadata server is reachable from GCE / Cloud Run.
BLOCKED_HOSTNAMES = frozenset({"metadata.google.internal", "metadata"})

# Link-local ranges. 169.254.169.254 is the metadata server; the whole range is
# reserved for link-local use and is never a legitimate test target.
BLOCKED_NETWORKS = (
    ipaddress.ip_network("169.254.0.0/16"),
    ipaddress.ip_network("fe80::/10"),
)

# Header the metadata server requires on every request. No legitimate test
# target needs it, so its presence is treated as an error rather than stripped.
METADATA_HEADER = "metadata-flavor"

MAX_REDIRECTS = 5


class BlockedTargetError(ValueError):
    """Raised when a request would be sent to a disallowed destination."""


def _is_blocked_address(address):
    """Returns True if an IP address falls inside a blocked network."""
    if isinstance(address, ipaddress.IPv6Address) and address.ipv4_mapped:
        address = address.ipv4_mapped
    return any(address in network for network in BLOCKED_NETWORKS)


def validate_target_url(url):
    """Checks that `url` may be requested by the worker.

    The URL must use http(s) and must not point at the metadata server, either
    directly (hostname or IP literal) or via a DNS name that resolves to a
    blocked address.

    Args:
        url: The absolute URL about to be requested.

    Raises:
        BlockedTargetError: If the URL is malformed or targets a blocked
            destination.
    """
    try:
        parts = urlsplit(url or "")
    except ValueError as e:
        raise BlockedTargetError(f"Invalid URL: {url!r}") from e

    if parts.scheme not in ("http", "https") or not parts.hostname:
        raise BlockedTargetError(
            f"Unsupported or malformed URL (expected http(s)://host/...): {url!r}"
        )

    hostname = parts.hostname.lower().rstrip(".")
    if hostname in BLOCKED_HOSTNAMES:
        raise BlockedTargetError(f"Requests to {hostname!r} are not allowed")

    try:
        addresses = {ipaddress.ip_address(hostname)}
    except ValueError:
        # Not an IP literal: resolve it so that DNS names aliasing a blocked
        # address are caught too. Resolution failures are left to the HTTP
        # client, which reports them as ordinary connection errors.
        try:
            default_port = 443 if parts.scheme == "https" else 80
            infos = socket.getaddrinfo(
                hostname, parts.port or default_port, proto=socket.IPPROTO_TCP
            )
        except (socket.gaierror, UnicodeError):
            return
        addresses = {ipaddress.ip_address(info[4][0]) for info in infos}

    for address in addresses:
        if _is_blocked_address(address):
            raise BlockedTargetError(
                f"Requests to {hostname!r} ({address}) are not allowed"
            )


def validate_headers(headers):
    """Rejects headers that only make sense when talking to the metadata server.

    Args:
        headers: Mapping of request headers (may be None).

    Raises:
        BlockedTargetError: If a ``Metadata-Flavor`` header is present.
    """
    for name in headers or {}:
        if str(name).lower() == METADATA_HEADER:
            raise BlockedTargetError(f"The {name!r} header is not allowed")


def send_request(method, url, json=None, headers=None, **kwargs):
    """Sends an HTTP request, validating the destination of every hop.

    Redirects are followed manually (up to MAX_REDIRECTS) so that each
    ``Location`` is checked before it is requested; a redirect to the metadata
    server is refused just like a direct request to it.

    Args:
        method: HTTP method.
        url: Target URL.
        json: Optional JSON-serialisable request body.
        headers: Optional request headers.
        **kwargs: Passed through to ``requests`` (e.g. ``timeout``).

    Returns:
        The final ``requests.Response``.

    Raises:
        BlockedTargetError: If the initial URL, any redirect target, or the
            headers are disallowed.
        requests.exceptions.TooManyRedirects: If the redirect chain is too long.
    """
    validate_headers(headers)
    validate_target_url(url)

    with requests.Session() as session:
        response = session.request(
            method, url, json=json, headers=headers, allow_redirects=False, **kwargs
        )
        redirects = 0
        while response.next is not None:
            redirects += 1
            if redirects > MAX_REDIRECTS:
                raise requests.exceptions.TooManyRedirects(
                    f"Exceeded {MAX_REDIRECTS} redirects", response=response
                )
            validate_target_url(response.next.url)
            response = session.send(response.next, allow_redirects=False, **kwargs)
        return response
