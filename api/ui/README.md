# Litmus UI

This repository hosts the user interface (UI) for Litmus, an AI-powered tool designed for streamlined testing and evaluation of HTTP requests and responses, especially those involving Large Language Models (LLMs). The UI empowers users to interact with the Litmus platform, manage test templates, and visualize test run results in an intuitive and insightful manner.

## Features

- **Test Creation and Management:**

  - Effortlessly create new test templates or modify existing ones through a user-friendly interface.
  - Define test cases with expected responses, incorporating placeholders for dynamic data substitution.
  - Craft pre-requests and post-requests to set up or validate the testing environment.
  - Define LLM prompts for AI-driven assessment of responses.

- **Test Run Submission:**

  - Submit test runs using defined templates, supplying unique run IDs for identification.
  - Monitor the progress of your test runs with real-time status updates.

- **Result Visualization and Analysis:**
  - View comprehensive details of each test run, including individual test case results.
  - Analyze request and response data, with options to filter by specific JSON paths for focused insights.
  - Explore historical data and compare runs to assess model performance over time.
  - Utilize interactive charts to visualize key metrics and identify trends.

## Technologies Used

- **Vue 3:** A progressive JavaScript framework for building user interfaces.
- **Vuetify 3:** A Material Design component framework for Vue.
- **Pinia:** A state management library for Vue, providing a centralized store for managing application data.
- **Vue Router:** A routing library for Vue, enabling navigation between different views.
- **Naive UI:** A Vue 3 component library providing a rich set of UI elements.
- **Chart.js:** A JavaScript charting library for creating interactive visualizations.
- **TypeScript:** A superset of JavaScript that adds static typing for improved code quality and maintainability.
- **SCSS:** A CSS preprocessor for writing modular and reusable styles.

## Getting Started

- Prerequisites:
  - Ensure you have a Litmus backend deployed and running. Refer to the Litmus backend documentation for instructions.
- Installation:
  - `npm install`
- Run Development Server:
  - `npm run dev`
- Build for Production:
  - `npm run build`

## Development

- Linting:
  - `npm run lint`
- Type Checking:
  - `npm run typecheck`

## Configuration

- The UI assumes the Litmus backend is running at the default location (`/`). If your backend is deployed at a different URL, you can configure the UI's base URL in the `vite.config.ts` file.

## Contributing

Contributions to the Litmus UI are welcome! Please refer to the contributing guidelines in the main Litmus repository.
