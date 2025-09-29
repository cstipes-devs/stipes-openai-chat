# Project Summaries for Context

## Digital Receipts Ingestion at Fetch Rewards

**Summary**  
We built a system to ingest digital orders from multiple data sources, parse store & product info, and transform them into *Fetch Receipts* which can be matched to user rewards. The pipeline includes filtering, validation, and enrichment across several services and data types.

---

### Supported Data Sources & Types

| Source      | Type / Integration Method                          |
|-------------|-----------------------------------------------------|
| Yahoo        | Email ingestion via Yahoo IMAP APIs (Go service)     |
| Outlook       | Email ingestion via Outlook email APIs (Go service)  |
| Gmail         | Email ingestion via Gmail APIs (Go service)          |
| Amazon         | JavaScript-based web scraping in a mobile webview     |
| Walmart        | JavaScript-based web scraping in a mobile webview     |
| Albertsons     | Direct API integration via Albertsons user/order APIs (Go service) |

---

### Technical Architecture & Components

1. **JS Web Scraping in Mobile WebView**
    - Runs in a client-side mobile webview environment
    - Executes JavaScript-based scraping logic to extract order data (e.g. order IDs, product details)
    - Communicates results to backend endpoints for post-processing

2. **Go Microservices for Email Ingestion**
    - Polls / fetches emails from inboxes (Yahoo, Outlook, Gmail)
    - Applies sender/address filters and lookback windows
    - Forwards raw email payloads to filtering & parsing layers

3. **Filtering & Order-Email Classification**
    - **Go** service: Uses a SageMaker model + heuristics to filter out non-order emails
    - **Python** service: Wraps another SageMaker model + heuristics for preprocessing/postprocessing of special cases (e.g. split orders, multi-part receipts)

4. **Order Storage & Approval Workflow**
    - **Go** microservice: Accepts parsed orders and stores them in **DynamoDB** (for persistence) and **Redis** (for caching / fast lookups)
    - Provides a UI or API endpoint for user review / approval before final submission to rewards processing

---

### Key Challenges & Solutions

- **Noise and non-order emails**: Leveraged ML + heuristics to reduce false positives
- **Heterogeneous data formats**: Created normalization layers to flatten differences across store, product, and receipt representations
- **Scalability**: Designed microservices to scale separately (email ingestion, filtering, storage)
- **Data consistency**: Used DynamoDB for persistent state, Redis for fast lookups, and idempotency checks to avoid duplications
- **User experience**: Built review/approval step so users can confirm or correct parsed data before reward processing

---

### Outcomes & Impact

- Successfully ingested **millions** of orders per week across all supported sources
- Reduced failure rate / errors in parsing by **X %** (you can insert your measured improvement)
- Improved user trust and satisfaction by giving users visibility and control over their parsed orders
- Enabled Fetch to process more digital receipts, increasing reward eligibility and retention

---

> *Note: This summary is available as context for AI agents to use when answering related queries or demonstrating your technical leadership in this space.*  