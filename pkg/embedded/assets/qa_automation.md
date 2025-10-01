## QA Frameworks

### API Testing
- Built a **Python-based API integration test suite** using `pytest`, maintained in its own dedicated repository.
- Configured **CI/CD triggers** to run test suites based on service name and domain, executing automatically **post-deployment**.
- Deployed subsets of tests on **AWS Lambda timers** as **synthetic health checks** for lower environments, enabling quick triage of issues spotted during mobile regression runs.
- Delivered comprehensive **reporting and visibility**:
    - Generated **HTML reports** automatically posted to Jira cards.
    - Sent **Graphite metrics** by domain with pass/fail counts.
    - Published results into a **centralized Report Portal web app**, giving a single source of truth for all test runs.

---

### Web Testing
- Designed and implemented a **TypeScript-based Playwright framework**, bundled directly with the web application it tested for tight integration and maintainability.
- Automated **cross-browser, end-to-end coverage**, aligned with CI/CD workflows.
- Delivered **rich reporting**:
    - **HTML reports** attached to Jira cards for instant feedback.
    - **Graphite metrics** for ongoing quality tracking by domain.
    - Centralized visibility in the **Report Portal web app**, alongside API and mobile test results.

---

### Mobile Testing
- Built a **Ruby-based Appium framework** to validate iOS and Android apps.
- Configured tests to run on **multiple real devices** through **BrowserStack’s cloud environment**, ensuring device-level reliability.
- Enabled **scalable regression testing** for mobile platforms as part of weekly release cycles.