# File Manager

The Litmus File Manager provides a centralized location for managing files that are referenced in your Litmus test cases or template definitions. This eliminates the need to embed large text content directly into your JSON payloads, making them more readable and easier to maintain.

![File Manager View](/img/file-manager.png)

## Benefits of Using the File Manager

- **Improved Readability:** Test cases and templates become more concise and easier to understand without large text blocks.
- **Version Control:** Track changes to your files independently, making it easier to identify and revert to previous versions.
- **Reduced Payload Size:** Smaller JSON payloads lead to faster processing and lower storage costs.
- **Reusability:** Easily reuse the same files across multiple test cases and templates.

## Uploading Files

1. **Access the File Manager:** Navigate to the "File Manager" page in the Litmus UI sidebar.
2. **Upload Files:**
   - Click or drag files to the designated upload area.
   - You can upload multiple files simultaneously.
3. **Confirm Upload:** Click the "Upload" button to initiate the file upload to your configured Google Cloud Storage bucket.

## Referencing Files in Test Cases and Templates

Once files are uploaded, you can reference them within your test cases and template definitions using the following format:

```
[FILE: your-file-name.txt]
```

**Example:**

Let's say you have a file named `sample_query.txt` in the File Manager containing the following text:

```
What is the capital of France?
```

In your test case definition, you can reference this file in the `query` field:

```json
{
  "query": "[FILE: sample_query.txt]",
  "response": "Paris"
}
```

When Litmus processes this test case, it will automatically replace the file reference with the actual content from `sample_query.txt`.

## Using the File Manager for Different Scenarios

### Test Cases

- **Input Data:** Store large text inputs, such as lengthy prompts or articles, in separate files and reference them in your test case definitions.
- **Golden Responses:** Maintain expected outputs for your test cases in files, making it easier to update and manage them.

### Template Definitions

- **Request Payloads:** Construct complex request payloads by referencing files containing specific sections of the payload. This is particularly useful for managing templates with extensive JSON structures.
- **LLM Evaluation Prompts:** Store your LLM evaluation prompts in separate files, keeping your template definitions cleaner and more focused.

## Important Notes

- Ensure your files are uploaded to the File Manager before referencing them in your test cases or templates.
- Use descriptive file names for easy identification.
- When referencing files, ensure the file path is correct, including any subfolders within your Google Cloud Storage bucket.
- The File Manager currently supports text-based files, such as `.txt` and `.json`.

By leveraging the Litmus File Manager, you can streamline your testing process, enhance the organization of your test data, and reduce the complexity of your JSON payloads, ultimately leading to more efficient and maintainable LLM testing workflows.
