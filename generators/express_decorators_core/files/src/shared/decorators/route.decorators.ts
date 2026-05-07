import { addRoute, getRoutes, type HttpMethod, type RouteMetadata } from './metadata';

export interface RouteOptions {
  path?: string;
  summary?: string;
  description?: string;
}

function buildRoute(method: HttpMethod, opts: string | RouteOptions | undefined, handlerName: string): RouteMetadata {
  if (typeof opts === 'string') {
    return { method, path: opts, handlerName };
  }
  return {
    method,
    path: opts?.path ?? '',
    handlerName,
    summary: opts?.summary,
    description: opts?.description,
  };
}

function makeRouteDecorator(method: HttpMethod) {
  return (pathOrOptions?: string | RouteOptions): MethodDecorator =>
    (target, propertyKey) => {
      addRoute(target, buildRoute(method, pathOrOptions, propertyKey as string));
    };
}

export const Get = makeRouteDecorator('get');
export const Post = makeRouteDecorator('post');
export const Put = makeRouteDecorator('put');
export const Patch = makeRouteDecorator('patch');
export const Delete = makeRouteDecorator('delete');

/**
 * Override the OpenAPI summary for a route after the route decorator was applied.
 */
export function Summary(summary: string): MethodDecorator {
  return (target, propertyKey) => {
    const routes = getRoutes(target);
    const route = routes.find((r) => r.handlerName === propertyKey);
    if (route) route.summary = summary;
  };
}

/**
 * Override the OpenAPI long description for a route.
 */
export function Description(description: string): MethodDecorator {
  return (target, propertyKey) => {
    const routes = getRoutes(target);
    const route = routes.find((r) => r.handlerName === propertyKey);
    if (route) route.description = description;
  };
}
