import { setRequiredHeaders } from './metadata';

/**
 * Mark headers required for the route. The router adapter installs a
 * pre-handler middleware that returns 400 if any are missing.
 */
export function RequiredHeaders(headers: string[]): MethodDecorator {
  return (target, propertyKey) => {
    setRequiredHeaders(target, propertyKey as string, headers);
  };
}
