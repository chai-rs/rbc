# Part 3: Testing Strategy

## 1. Why you selected it?

For my selected scenario is order with base price of 100, country code "TH", customer type "VIP", and is the customer's first order.  

## 2. Why scenario it covers?

This scenario covers the application of all pricing rules in the chain: tax calculation, first-order discount, customer discount, and rounding. It also tests the interaction between these rules and ensures that they are applied in the correct order, producing the expected final price.

## 3. What kind of bug it would catch?

This scenario would catch bugs related to the correct application of tax rates, discounts, and rounding. It would also reveal issues with the order of operations, such as applying discounts before taxes or vice versa. Additionally, it could uncover edge cases where the price might drop to zero or below, ensuring that the `ZeroPriceAction` is triggered correctly. Finally, it would validate that the final price is rounded up to the specified precision.

## 4. Are there any cases you intentionally did not test? Why?

I intentionally did not test scenarios with unsupported country codes or customer types, as these would be expected to return errors based on the current implementation. While it's important to eventually cover these cases, my initial focus is on validating the core functionality of the pricing rules and their interactions. Once the main logic is confirmed to be working correctly, I would then expand the test suite to include edge cases and error handling scenarios.

## 5. How would make you confident enough to release change to production?

To gain confidence in releasing the change to production, I would ensure that the test suite has comprehensive coverage of all critical paths and edge cases.
