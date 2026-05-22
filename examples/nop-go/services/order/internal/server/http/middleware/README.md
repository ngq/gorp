# middleware

Service-local HTTP middleware lives here when needed.

Keep middleware local to `order-service` unless it becomes clearly stable and shared. Do not move business rules or framework bootstrap details into this directory.
