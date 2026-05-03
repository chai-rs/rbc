# Part 4: External Payment API

The pricing function will be used before calling an external payment provider.

The provider can be slow, fail, or return an unclear result after a timeout.

In some cases, the provider may have processed the request, but your system did not receive the response.

## Assignments

### 1. What boundary/interface would you put between your code and the payment provider?

I would put the `PaymentGateway` interface between application and the external provider for testability. This PaymentGateway have responsiblity to enqueue the task to the worker with idempotency-key which generated and store at the application layer. for the job it used to call the provider and update the status of the order in the database another thing is used to handle the retry with backoff-jitter + timeout for the provider call but if it really failed after retrying the job, it will put the message to DLQ for investigation, analytics and retry later if needed. This way we can make sure that the same order is not charged twice and also we can handle the failure of the provider gracefully without affecting the user experience.

### 2. How would you make sure the same order is not charged twice?

To ensure that the same order is not charged twice, I would implement an idempotency mechanism.

### 3. What would you store before calling the provider, and why?

Before calling the provider, I would store the order details along with a unique idempotency key in a database. This allows us to track the status of each order and ensure that if a request is retried (due to a timeout or failure), we can check if the order has already been processed using the idempotency key. If it has, we can return the previous result instead of making another call to the provider, thus preventing duplicate charges.

### 4. What would you unit test with a fake provider?

With a fake provider when consumed, I would unit test the following scenarios:
- Successful payment processing.
- Handling of provider timeouts.
- Retry logic with backoff and jitter.
- Idempotency mechanism to prevent duplicate charges.
- Proper handling of failed payments and moving them to the DLQ.
