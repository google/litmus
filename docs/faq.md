## Frequently Asked Questions (FAQ)

**What is Litmus?**
Litmus is a tool for quickly building and testing LLMs by providing a user interface for creating and managing test templates, submitting test runs and analyzing test results. It also provides an optional proxy service for logging LLM interactions which can then be analyzed.

**How can I use Litmus?**
There are two ways to use Litmus:

- **Use the provided CLI:** This is the easiest way to set up Litmus.
- **Manual Setup:** This allows for more customization.

See the [README](https://github.com/google/litmus/) for more details.

**What are the benefits of using Litmus?**
Litmus offers several benefits for testing and evaluating LLMs:

- **Automated Test Execution:** Streamlines the testing process by automating the execution of test cases based on predefined templates.
- **Flexible Test Templates:** Allows you to define and manage test templates to specify the structure and parameters of your tests, enabling customization and reusability.
- **User-Friendly Web Interface:** Provides an intuitive and visually appealing web interface to interact with the Litmus platform, simplifying test creation, submission, and analysis.
- **Detailed Results:** Offers comprehensive insights into the status, progress, and detailed results of your test runs, facilitating thorough analysis and identification of potential issues.
- **Advanced Filtering:** Enables you to filter responses from test runs based on specific JSON paths, focusing your analysis on specific aspects of the LLM's output.
- **Performance Monitoring:** Tracks the performance of your LLM responses using AI-powered evaluation, allowing you to identify areas for improvement and optimize your models.
- **LLM Evaluation with Customizable Prompts:** Leverages LLMs to compare actual responses with expected (golden) responses, utilizing customizable prompts to tailor the evaluation to your specific needs.
- **Proxy Service for Enhanced LLM Monitoring:** Provides an optional proxy service to capture and analyze LLM interactions in greater detail, giving you a comprehensive understanding of your LLM usage patterns.
- **Cloud Integration:** Seamlessly integrates with Google Cloud Platform (Firestore, Cloud Run, BigQuery) for efficient data storage, execution, and analysis, leveraging the scalability and reliability of cloud services.
- **Quick Deployment:** Facilitates streamlined setup through a deployment tool, simplifying the process of getting Litmus up and running in your environment.

**What LLMs does Litmus support?**
Currently, Litmus supports Google's [Gemini](https://ai.google.dev/tutorials/get-started-gemini) family of models. The default model can be configured in [settings.py](https://github.com/google/litmus/blob/main/api/util/settings.py).

**How can I get help?**
Please refer to the [README](https://github.com/google/litmus/) for the different ways to get help.

**How can I contribute?**
Contributions to Litmus are welcome! Please refer to the [CONTRIBUTING](https://github.com/google/litmus/blob/main/CONTRIBUTING.md) file for guidelines and instructions on how to contribute.
