import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  base: "/litmus/",
  outDir: "dist",
  cleanUrls: true,
  title: "google/litmus",
  titleTemplate: ":title | google/litmus",
  description: "LLM testing and evaluation tool",
  head: [
    [
      "link",
      {
        rel: "apple-touch-icon",
        sizes: "180x180",
        href: "/litmus/img/favicons/apple-touch-icon.png",
      },
    ],
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        sizes: "32x32",
        href: "/litmus/img/favicons/favicon-32x32.png",
      },
    ],
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        sizes: "16x16",
        href: "/litmus/img/favicons/favicon-16x16.png",
      },
    ],
    [
      "link",
      {
        rel: "mask-icon",
        href: "/litmus/img/favicons/safari-pinned-tab.svg",
        color: "#3a0839",
      },
    ],
    [
      "link",
      { rel: "shortcut icon", href: "/litmus/img/favicons/favicon.ico" },
    ],
    ["meta", { name: "og:image", content: "/litmus/img/og-image.png" }],
    ["meta", { name: "twitter:image", content: "/litmus/img/og-image.png" }],
  ],
  themeConfig: {
    logo: "/img/logo.svg",
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Home", link: "/" },
      { text: "Docs", link: "/getting-started" },
    ],

    sidebar: {
      "/": [
        {
          text: "Docs",
          items: [
            { text: "Getting Started", link: "/getting-started" },
            { text: "Setup", link: "/setup" },
            { text: "API Reference", link: "/api" },
            { text: "CLI Usage", link: "/cli" },
            { text: "Proxy Usage", link: "/proxy" },
            { text: "Contribution Guide", link: "/contribution" },
          ],
        },
        {
          text: "Template Types",
          items: [
            { text: "Test Run", link: "/template-test-run" },
            { text: "Test Mission", link: "/template-test-mission" },
          ],
        },
        {
          text: "Using Litmus",
          items: [
            { text: "Adding Templates", link: "/ui-adding-templates" },
            { text: "Starting a Test Run", link: "/ui-start-test-run" },
            { text: "Test Run Analysis", link: "/ui-test-run-analysis" },
            { text: "Test Run Comparison", link: "/ui-test-run-comparison" },
          ],
        },
        {
          text: "FAQ",
          link: "/faq",
          items: [{ text: "Known Issues", link: "/known-issues" }],
        },
      ],
    },

    socialLinks: [{ icon: "github", link: "https://github.com/google/litmus" }],

    editLink: {
      pattern: "https://github.com/google/litmus/blob/main/docs/:path",
    },

    footer: {
      message:
        "Disclaimer: This is not an officially supported Google product.",
    },

    search: {
      provider: "local",
    },
  },
});
