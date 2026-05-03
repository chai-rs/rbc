# Part 2: Refactoring/ Code Review

The AI-generated code provided with the assignment is reviewed below.

```javascript
function calculateFinalPrice(order) {
  let price = order.basePrice;

  if (order.countryCode === "TH") {
    price = price + price * 0.07;
  } else {
    price = price + price * 0.2;
  }

  if (order.customerType === "VIP") {
    price = price - 100; 
  }
  
  if (order.isFirstOrder) {
    price = price * 0.9;
  }

  return price.toFixed(2);
}
```

## Assignments

### 1. What bugs or specification mismatches do you see?

1. **Coupling** logic to specific country codes and customer types, making it hard to extend or maintain.
2. **Hardcoded** tax rates and discounts, which should ideally come from a configuration or environment variables.
3. **Returned price is a string** due to `toFixed(2)`, which may not be ideal for further calculations or comparisons.
4. **No error handling** for invalid inputs (e.g., negative base price, unsupported country codes, or customer types).
5. **Lack of extensibility** for new business rules, such as additional customer types or promotional discounts.
6. **Fixed order of operations** that may not be flexible for different pricing strategies (e.g., applying discounts before taxes).

### 2. What production risks do you see?

1. **Inflexibility**: The current implementation is rigid and would require code changes to accommodate new rules, which increases the risk of introducing bugs.
2. **Scalability issues**: As the number of rules grows, the function will become increasingly complex and difficult to manage.
3. **Configuration management**: Hardcoded values make it difficult to adjust pricing strategies without modifying the codebase, which can lead to errors and downtime.
4. **Data integrity**: Returning a string instead of a numeric value can lead to issues in downstream processes that expect a number, such as further calculations or database storage.
5. **Lack of validation**: The function does not validate inputs, which could lead to incorrect pricing or system crashes if invalid data is passed.

### 3. How would you refactor it?

1. **Introduce a configuration system** to manage tax rates and discounts, allowing for easier updates without code changes.
2. **Decouple business logic** by implementing a more modular design, such as using a chain of responsibility pattern for applying pricing rules.
3. **Return a numeric value** instead of a string to maintain data integrity and allow for further calculations.
4. **Add error handling** to validate inputs and handle edge cases gracefully.
5. **Make the order of operations flexible** by allowing rules to be applied in a configurable sequence, rather than a fixed order.

### 4. What is the first test you would write to validate your refactor, and why?

The first test I would write is a unittest for the `Calculator.Calculate` method to ensure that it correctly applies the rules in the expected order, produces the correct final price for a given set of inputs and handle zero or negative price scenario properly.
