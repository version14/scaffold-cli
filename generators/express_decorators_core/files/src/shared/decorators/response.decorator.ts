import type { ZodSchema } from 'zod';
import { addResponse } from './metadata';

/**
 * Declare an OpenAPI response for a route. Multiple ApiResponse decorators
 * can be stacked on the same handler to document several status codes.
 */
export function ApiResponse(status: number, description: string, schema?: ZodSchema): MethodDecorator {
  return (target, propertyKey) => {
    addResponse(target, propertyKey as string, { status, description, schema });
  };
}
