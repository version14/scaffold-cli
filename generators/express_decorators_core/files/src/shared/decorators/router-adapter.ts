import type { ZodSchema } from 'zod';
import type { HttpMethod } from './metadata';

export interface RouteRegistration {
  method: HttpMethod;
  path: string;
  handler: (...args: unknown[]) => unknown;
  validation: {
    body?: ZodSchema;
    params?: ZodSchema;
    query?: ZodSchema;
  };
  requiredHeaders: string[];
  isProtected: boolean;
}

/**
 * Framework-agnostic router contract. Implementations translate a
 * RouteRegistration into framework-native middleware + handler wiring.
 *
 * Express adapter is shipped by default. To support Fastify or another
 * framework, implement this interface and pass an instance to DecoratorRouter.
 */
export interface RouterAdapter<NativeRouter = unknown> {
  register(registration: RouteRegistration): void;
  build(): NativeRouter;
}
