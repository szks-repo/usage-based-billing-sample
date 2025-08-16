# usage-based-billing-sample

```sh
# run RabbitMq Container
docker compose up -d

# run apiServer, consumerWorker
mise run run*
```

## Roadmap

| Phase | Status | Description |
| :--- | :---: | :--- |
| **1. Reliable Usage Data Collection** | ✅ | **Store Usage Logs in S3 (Parquet Format):** Persist API usage logs to Amazon S3 in Parquet format. This establishes a cost-effective, scalable, and analyzable "source of truth" for all billing data. |
| **2. Near Real-time Usage Aggregation** | ⬜️ | **Persist Usage Data from Worker to DB:** Implement an asynchronous worker to process logs, calculate usage, and store the aggregated data in a database. This provides users with near real-time access to their usage information without impacting API performance. |
| **3. Data Integrity and Reconciliation** | ⬜️ | **Implement Reconciliation Process:** Create a batch process to compare the Parquet logs (the source of truth) with the aggregated usage data in the database. This ensures billing accuracy by detecting and correcting any discrepancies caused by network issues or worker failures. |

