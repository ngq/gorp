# middleware

Service-local HTTP middleware lives here when needed.

Use this directory only for middleware owned by `product-service`. If the behavior is business-specific, keep it near the service. If it becomes stable and shared across services, evaluate extracting it later instead of defaulting to a shared directory too early.
