import { setProtected } from './metadata';

/**
 * Marks a route as protected. The router adapter is responsible for installing
 * the auth middleware and OpenAPI security scheme. Without an auth middleware
 * registered on the adapter, the decorator only affects the OpenAPI spec.
 */
export function Auth(): MethodDecorator {
  return (target, propertyKey) => {
    setProtected(target, propertyKey as string);
  };
}
