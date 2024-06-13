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

from vertexai.language_models._language_models import GroundingCitation
import re
import datetime


def cleanup_json(x: str):
    """Takes a string with supposed JSON content and eliminates excess characters at beginning and end.

    Args:
        x (str): String containing a JSON structure. May contain leading and trailing characters outside of JSON.

    Returns:
        str: Retuns the substring representig the JSON structure. Returns None if JSON could not be found.
    """
    y = x.replace("'", "")
    startBrace = y.find("{")
    if startBrace == -1:
        return None
    endBrace = y.rfind("}")
    if endBrace == -1:
        return None
    if startBrace > endBrace:
        return None
    return y[startBrace : endBrace + 1].replace("\n", " ")


def strip_references(statement: str):
    """Strips reference notation elements (e.g. [3]) from a string.

    Args:
        statement (str): The string that could contain references

    Returns:
        The string with any references removed.
    """
    out = ""
    p = 0
    while True:
        if p == len(statement):
            break
        q = statement.find("[", p)
        out += statement[p:q]
        if q < 0:
            break
        else:
            v = statement.find("]", q + 1)
            if v < 0:
                break
            p = v + 1
    return out


def find_citations(text_with_citations) -> list[tuple]:
    # Find all citations in the text
    citations = []
    p = 0
    q = -1
    while True:
        # find a citation
        p = text_with_citations.find("[", q + 1)
        if p < 0:  # no (more) citations
            break
        # find the end of this particular citation
        q = text_with_citations.find("]", p)
        if q < 0:  # in case we have a missing closing bracket
            break
        # now iterate
        p += 1
        while p < q:
            r = p
            if text_with_citations[r].isdigit():
                while r <= q and text_with_citations[r].isdigit():
                    r += 1
                citations.append((p, r, int(text_with_citations[p:r])))
                while r < q and not text_with_citations[r].isdigit():
                    r += 1
                p = r
            else:
                break
        p = q + 1
    # now we have a list of citation locations
    citations.reverse()
    return citations


def replace_citation(
    text_with_citations: str, startpos: int, endpos: int, newValue: int
) -> str:
    return (
        text_with_citations[0:startpos] + str(newValue) + text_with_citations[endpos:]
    )


def renumber_citations(search_result, offset: int):
    # Find all citations in the text
    citations = find_citations(search_result)
    searchresponse = search_result
    # replace all citations with new values
    for cit_start, cit_end, cit_value in citations:
        cit_value += offset
        searchresponse = replace_citation(searchresponse, cit_start, cit_end, cit_value)
    return searchresponse


def insert_citations(response: str, citations: [GroundingCitation]):
    """insert citations from citations into text"""
    cits = [
        {"start": c.start_index, "end": c.end_index, "id": i}
        for i, c in enumerate(citations, start=1)
    ]
    cits.sort(key=lambda x: x["end"], reverse=True)
    for c in cits:
        end = c["end"] + 1
        response = response[:end] + f"[{c['id']}]" + response[end:]
    return response


def get_doc_cit(doc):
    n = doc["name"]
    p = n.find("]")
    return int(n[1:p])


def renumber_docs(search_result, offset: int):
    for doc in search_result.get("documents", []):
        n = doc["name"]
        p = n.find("]")
        doc["name"] = f"[{str(int(n[1:p])+offset)}{n[p:]}"


def compact_docs(documents: list[dict], citations: list[tuple]):
    """Compacts the list of documents by removing all documents not mentioned by citations.

    Args:
        documents (list[dict]): List of documents as returned by e.g., search_engine_summary
        citations (list[tuple]): List of citations as returned by find_citations

    Returns:
        No return but manipulates the list of documents (reduces it)
    """
    # set of unique citation ids
    cits = set([v for _, _, v in citations])
    for i in range(len(documents) - 1, -1, -1):
        if not get_doc_cit(documents[i]) in cits:
            documents.pop(i)


def extract_article_id(url):
    """Extracts the 8-digit number from a Blick.ch URL.

    Args:
    url: The URL string.

    Returns:
    The extracted number as a string, or None if no match is found.
    """
    id_pattern = r"(\d{8})(?=\D|$)"

    # Search for the pattern in each URL and extract the ID numbers.
    matched_ids = re.findall(id_pattern, url)

    # Sometimes, the pattern might match multiple IDs; we assume the first one is the correct ID.
    extracted_id = matched_ids[0] if matched_ids else None

    return extracted_id


def date_str_to_unix_int(date_str):
    """Converts a date string in YYYY-MM-DD format to a Unix timestamp integer.

    Args:
      date_str: The date string to convert.

    Returns:
      The Unix timestamp as an integer.
    """
    try:
        datetime_obj = datetime.datetime.strptime(date_str, "%Y-%m-%d")
        unix_timestamp = int(datetime_obj.timestamp())
        return unix_timestamp
    except ValueError:
        return None  # Or raise an exception if preferred
