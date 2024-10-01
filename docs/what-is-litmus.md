# What is Litmus?

Litmus is a comprehensive, AI-powered testing and evaluation platform designed to empower developers building and deploying LLM-powered applications. It provides the tools and insights needed to ensure the reliability, performance, and safety of generative AI solutions.

<video controls="controls" src="/video/Litmus.mp4" />

## The Challenges of LLM Testing

Testing LLM-powered applications presents unique challenges compared to traditional software testing:

- **Non-deterministic Outputs:** LLMs can produce varied outputs for identical inputs, making traditional testing methods unreliable.
- **Difficulty in Defining "Correctness":** Evaluating LLM outputs often involves subjective judgments and nuanced assessments of quality, relevance, and context.
- **Potential for Biases, Hallucinations, and Safety Issues:** LLMs can inherit biases from training data, generate factually incorrect information, or produce harmful content, requiring careful scrutiny and mitigation strategies.

## Litmus: A Comprehensive Solution

![Litmus LLM Testing](/img/litmus.png)

Litmus addresses these challenges with a suite of features designed to streamline the testing and evaluation process:

**1. Flexible Test Templates:**

- Create reusable test templates for both single-turn and multi-turn interactions.
- Define diverse scenarios and customize parameters and inputs.
- Specify evaluation criteria and metrics tailored to your application's needs.

**2. Automated Test Execution:**

- Automate the execution of test cases using your defined templates and test data.
- Effortlessly submit test runs and monitor progress with real-time status updates.

**3. Detailed Result Analysis:**

- Visualize comprehensive test results with clear pass/fail indicators and AI-driven assessments.
- Gain in-depth insights into model performance, identify areas for improvement, and track progress over time.
- Compare different test runs to analyze how your model performs across various versions or configurations.
- Leverage interactive charts and filter results to focus on specific metrics or request/response patterns.

**4. Diverse LLM Evaluation Methods:**

- **Custom LLM Evaluation:** Employ a separate LLM to assess responses against golden answers, using customizable prompts to tailor evaluations.
- **Ragas Evaluation:** Utilize a comprehensive set of Ragas metrics to assess answer relevancy, context recall, precision, harmfulness, and similarity to reference answers.
- **DeepEval Evaluation:** Leverage DeepEval's LLM-powered metrics to delve deeper into aspects like faithfulness, contextual relevance, hallucination, bias, and toxicity.

**5. Proxy Service for Enhanced Monitoring:**

- Gain deeper insights into LLM usage patterns through a dedicated proxy service.
- Capture detailed request and response logs for analysis, debugging, and optimization.
- Explore proxy data in the Litmus UI to understand how your LLMs are being utilized and identify areas for refinement.

**6. Cloud Integration:**

- Seamlessly integrate with Google Cloud Platform services, including Firestore for data storage, Cloud Run for job execution, BigQuery for proxy log analysis, and Vertex AI for accessing powerful LLMs.
- Benefit from the scalability, reliability, and security of Google Cloud.

**7. Quick and Easy Deployment:**

- Effortlessly set up Litmus using the Litmus CLI tool, which automates the deployment process and configures necessary resources.
- Alternatively, follow the manual setup guide for more customized deployment options.

## Benefits of Using Litmus

Litmus empowers GenAI developers with:

- **Increased Confidence:** Build robust LLM applications with comprehensive testing and AI-powered evaluations.
- **Faster Development Cycles:** Streamline testing workflows with automated execution and intuitive result analysis.
- **Reduced Risk:** Identify and mitigate potential issues like biases and hallucinations early in development.
- **Improved Performance:** Track LLM performance metrics and optimize your applications for efficiency.

## Getting Started with Litmus

Ready to take your GenAI development to the next level? Visit our [Getting Started](/getting-started) page to get started with Litmus today!
