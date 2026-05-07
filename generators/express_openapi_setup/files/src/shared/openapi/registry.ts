import { extendZodWithOpenApi, OpenAPIRegistry } from '@asteasolutions/zod-to-openapi';
import { z } from 'zod';

extendZodWithOpenApi(z);

/**
 * Construct a fresh OpenAPI registry pre-populated with a BearerAuth security
 * scheme. Each call returns a new registry — this avoids global mutable state
 * and lets tests run in isolation.
 *
 * Production code typically wants the shared registry below.
 */
export function createRegistry(): OpenAPIRegistry {
  const registry = new OpenAPIRegistry();
  registry.registerComponent('securitySchemes', 'BearerAuth', {
    type: 'http',
    scheme: 'bearer',
    bearerFormat: 'JWT',
  });
  return registry;
}

let shared: OpenAPIRegistry | undefined;

/**
 * Lazily-initialised shared registry for production code that wants a single
 * collection point for every OpenAPI definition. Tests should prefer
 * createRegistry() to avoid leaking definitions between cases.
 */
export function getSharedRegistry(): OpenAPIRegistry {
  if (!shared) shared = createRegistry();
  return shared;
}

export function resetSharedRegistry(): void {
  shared = undefined;
}

export { z };
