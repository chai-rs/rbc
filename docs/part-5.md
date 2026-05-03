# Part 5: Production Incident

The pricing logic was wrong in production for 2 hours.
Some customers were overcharged.

1 What do you do first?

First, I would immediately stop the deployment or roll back to the previous stable version to prevent further customers from being affected. Then, I would assess the scope of the issue by checking logs and monitoring systems to identify how many customers were impacted and the extent of the overcharging.

2 Who do you inform?

I would inform the relevant stakeholders, including the customer support team, finance team, and management. It's crucial to keep everyone informed about the situation, the steps being taken to resolve it, and the expected timeline for a fix.

3 How do you identify affected orders?

I would query the order and payment databases to identify all orders processed during the window when the pricing logic was incorrect.

4 How do you fix the issue safely?

To fix the issue safely, I would first correct the pricing logic in the codebase and thoroughly test it in a SIT and UAT environment together with the QA team to ensure the bug is resolved. Once the fix is verified, I would deploy it to production with monitoring in place to ensure that the issue is resolved and does not recur. Additionally, I would work with the customer support and finance teams to identify affected customers and process refunds or adjustments as necessary.

5 How do you prevent it from happening again?

To prevent this from happening again, I would implement a more robust testing strategy that includes comprehensive unit tests, integration tests, and end-to-end tests for the pricing logic. I would also set up monitoring and alerting systems to detect anomalies in pricing or order processing in real-time. Additionally, I would consider implementing a feature flags and circuit breaker pattern for the pricing logic to prevent cascading failures in case of issues.

6 How would you have deployed this change to reduce the risk of this happening?

To reduce the risk of this happening, I would have deployed the change using a canary deployment strategy. This involves rolling out the new pricing logic to a small percentage of users first and monitoring the impact before fully deploying it to all users. This way, if there are any issues with the new logic, it would only affect a small subset of customers, allowing for quicker detection and rollback if necessary. Additionally, I would have implemented feature flags to enable or disable the new pricing logic without requiring a full deployment, providing an extra layer of control in case of issues.
