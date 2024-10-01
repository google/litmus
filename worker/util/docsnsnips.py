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
        str: Returns the substring representing the JSON structure. Returns None if JSON could not be found.
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
    """Strips reference notation elements (e.g., [3]) from a string.

    Args:
        statement (str): The string that could contain references

    Returns:
        str: The string with any references removed.
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
    """Finds all citations in the text and returns their locations and values.

    Args:
        text_with_citations (str): Text potentially containing citations in the form [number].

    Returns:
        list[tuple]: A list of tuples, each representing a citation.
                     Each tuple contains:
                         - start position (int): The starting index of the citation in the text.
                         - end position (int): The ending index of the citation in the text.
                         - citation value (int): The numerical value of the citation.
    """

    # Find all citations in the text
    citations = []
    p = 0
    q = -1
    while True:
        # Find a citation
        p = text_with_citations.find("[", q + 1)
        if p < 0:  # no (more) citations
            break
        # Find the end of this particular citation
        q = text_with_citations.find("]", p)
        if q < 0:  # In case we have a missing closing bracket
            break
        # Now iterate
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
    # Now we have a list of citation locations
    citations.reverse()
    return citations


def replace_citation(
    text_with_citations: str, startpos: int, endpos: int, newValue: int
) -> str:
    """Replaces a citation in a string with a new value.

    Args:
        text_with_citations (str): The text containing the citation to be replaced.
        startpos (int): The starting index of the citation.
        endpos (int): The ending index of the citation.
        newValue (int): The new value to replace the existing citation with.

    Returns:
        str: The updated string with the replaced citation.
    """
    return (
        text_with_citations[0:startpos] + str(newValue) + text_with_citations[endpos:]
    )


def renumber_citations(search_result, offset: int):
    """Renumbers the citations in a search result by a given offset.

    Args:
        search_result (str): The search result text potentially containing citations.
        offset (int): The amount to add to each citation number.

    Returns:
        str: The search result with renumbered citations.
    """
    # Find all citations in the text
    citations = find_citations(search_result)
    searchresponse = search_result
    # Replace all citations with new values
    for cit_start, cit_end, cit_value in citations:
        cit_value += offset
        searchresponse = replace_citation(searchresponse, cit_start, cit_end, cit_value)
    return searchresponse


def insert_citations(response: str, citations: [GroundingCitation]):
    """Inserts citations from a list of GroundingCitation objects into the text.

    Args:
        response (str): The text where citations should be inserted.
        citations (list[GroundingCitation]): A list of GroundingCitation objects from an LLM response.

    Returns:
        str: The text with citations inserted in the form [number] at their corresponding locations.
    """
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
    """Extracts the citation number from a document name in the search results.

    Args:
        doc (dict): A document dictionary from the search results, containing a 'name' field.

    Returns:
        int: The citation number extracted from the document name.
    """
    n = doc["name"]
    p = n.find("]")
    return int(n[1:p])


def renumber_docs(search_result, offset: int):
    """Renumbers the citations within document names in a search result.

    Args:
        search_result (dict): The search result potentially containing documents with citations in their names.
        offset (int): The amount to add to each citation number within document names.
    """
    for doc in search_result.get("documents", []):
        n = doc["name"]
        p = n.find("]")
        doc["name"] = f"[{str(int(n[1:p])+offset)}{n[p:]}"


def compact_docs(documents: list[dict], citations: list[tuple]):
    """Compacts the list of documents by removing documents not referenced in the citations.

    Args:
        documents (list[dict]): List of documents from search results, each with a 'name' field containing a citation.
        citations (list[tuple]): List of citations found in the text.

    This function modifies the `documents` list in-place, removing documents that are not cited.
    """
    # Set of unique citation IDs
    cits = set([v for _, _, v in citations])
    for i in range(len(documents) - 1, -1, -1):
        if get_doc_cit(documents[i]) not in cits:
            documents.pop(i)


def extract_article_id(url):
    """Extracts the 8-digit number from a Blick.ch URL.

    Args:
        url (str): The URL string.

    Returns:
        str or None: The extracted 8-digit number as a string, or None if no match is found.
    """
    id_pattern = r"(\d{8})(?=\D|$)"

    # Search for the pattern in the URL and extract ID numbers.
    matched_ids = re.findall(id_pattern, url)

    # Return the first matched ID, or None if no match is found.
    extracted_id = matched_ids[0] if matched_ids else None

    return extracted_id


def date_str_to_unix_int(date_str):
    """Converts a date string in YYYY-MM-DD format to a Unix timestamp integer.

    Args:
        date_str (str): The date string to convert.

    Returns:
        int or None: The Unix timestamp as an integer, or None if the input is invalid.
    """
    try:
        datetime_obj = datetime.datetime.strptime(date_str, "%Y-%m-%d")
        unix_timestamp = int(datetime_obj.timestamp())
        return unix_timestamp
    except ValueError:
        return None
