# Part 7: Code Styles and Trade-offs

For your implementation in [Part 1](docs/part-1.md)

## Assignments

### 1. How did you structure your code?

1. I separated the code into different packages based on their responsibilities, such as `order`, `pricerule`, and `calculator`. This keeps the code organized, easier to understand, and more maintainable. The `order` package contains the `Order` domain model and related logic.

2. The `order/customer.go` file is a slight exception to the package convention. Ideally, I would place `CustomerType` in a separate `customer` package. However, because this is a small project, I kept it inside the `order` package to avoid creating too many packages too early. In production, I would move it to its own package for better separation of concerns and long-term maintainability.

### 2. Would you describe it as OOP, functional, or a mix?

I would describe it as a mix of OOP and functional programming. The design uses OOP principles by defining domain models, such as `Order`, and encapsulating related data and behavior. Each pricing rule implements the same `PriceRule` interface, so the `Calculator` can work with different rule implementations without depending on their concrete types.

The rule execution also follows the Chain of Responsibility pattern. Each `PriceRule` receives the current state of the calculation, applies its own logic, and passes the updated price to the next rule in the chain. This keeps each rule small, focused, and easy to replace or reorder. It also makes the design open for extension because a new rule can be added to the chain without changing the existing rules.

At the same time, the pricing flow has a functional programming style. Each rule behaves like a transformation: it takes an `Order` and the current price as input, then returns a new price as output. The `Calculator` composes these transformations into a pipeline. This makes the pricing logic predictable, easier to test, and easier to reason about because each rule has a clear input and output.

### 3. Why did you choose this approach?

I chose this approach because it provides a good balance between flexibility, maintainability, and testability. The modular design allows for easy addition of new pricing rules without modifying existing code, which reduces the risk of introducing bugs.

### 4. If the pricing rules become much more complex, how would your approach evolve?

If the pricing rules become more complex than the current implementation can handle, I would evolve the design gradually. I would start with a decision table, which is a structured way to represent business rules as data. At first, I would query or load all active rules, loop through them, check which rules match the order, and apply the matching price rules in priority order.

As the number of pricing rules grows, I would avoid evaluating every rule for every order. Instead, I would query only the candidate rules based on fields such as country code, customer type, order amount, active date range, and rule priority. To reduce database load and improve performance, I would cache the active rules or the query results, then refresh the cache when rules are updated.

If the rules become even more complex, with many overlapping conditions, exclusions, priorities, or business-managed changes, I would consider using a dedicated rule engine. I would not start with a rule engine immediately because it adds operational and technical complexity before the project needs it.
