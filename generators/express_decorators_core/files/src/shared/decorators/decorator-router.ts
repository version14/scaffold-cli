import {
  getController,
  getRequiredHeaders,
  getResponses,
  getRoutes,
  getValidation,
  isProtected,
  type ControllerMetadata,
  type ResponseMetadata,
  type RouteMetadata,
  type ValidationMetadata,
} from './metadata';
import type { RouteRegistration, RouterAdapter } from './router-adapter';

export interface RegisteredRoute {
  controller: ControllerMetadata;
  route: RouteMetadata;
  validation: ValidationMetadata;
  responses: ResponseMetadata[];
  protected: boolean;
  fullPath: string;
}

function joinPath(prefix: string, path: string): string {
  const left = prefix.endsWith('/') ? prefix.slice(0, -1) : prefix;
  // A bare '/' or '' path means "the prefix itself" — emit just the prefix
  // (no trailing slash) so OpenAPI path keys are stable.
  if (path === '' || path === '/') return left || '/';
  const right = path.startsWith('/') ? path : `/${path}`;
  return `${left}${right}`;
}

/**
 * DecoratorRouter walks a controller instance, reads its decorator metadata,
 * registers each route on a RouterAdapter, and exposes the registered routes
 * for downstream consumers (notably the OpenAPI spec generator).
 */
export class DecoratorRouter<NativeRouter = unknown> {
  private readonly registered: RegisteredRoute[] = [];

  constructor(private readonly adapter: RouterAdapter<NativeRouter>) {}

  /**
   * Bind every decorated route on `instance` to the underlying adapter.
   * Returns this for fluent chaining.
   */
  registerController(instance: object): this {
    const proto = Object.getPrototypeOf(instance) as object;
    const controller = getController(proto);
    if (!controller) {
      throw new Error('DecoratorRouter: target is missing the @Controller decorator');
    }

    const routes = getRoutes(proto);
    for (const route of routes) {
      const handler = (instance as Record<string, unknown>)[route.handlerName];
      if (typeof handler !== 'function') {
        throw new Error(`DecoratorRouter: handler "${route.handlerName}" is not callable`);
      }

      const validation = getValidation(proto, route.handlerName);
      const responses = getResponses(proto, route.handlerName);
      const protectedRoute = isProtected(proto, route.handlerName);
      const requiredHeaders = getRequiredHeaders(proto, route.handlerName);
      const fullPath = joinPath(controller.prefix, route.path);

      const registration: RouteRegistration = {
        method: route.method,
        path: fullPath,
        handler: (handler as (...args: unknown[]) => unknown).bind(instance),
        validation,
        requiredHeaders,
        isProtected: protectedRoute,
      };

      this.adapter.register(registration);

      this.registered.push({
        controller,
        route,
        validation,
        responses,
        protected: protectedRoute,
        fullPath,
      });
    }

    return this;
  }

  /**
   * The native router (Express Router, Fastify instance, ...).
   */
  build(): NativeRouter {
    return this.adapter.build();
  }

  /**
   * Snapshot of every route the router has bound. Consumed by the OpenAPI
   * generator to emit a spec that exactly matches what is wired.
   */
  routes(): readonly RegisteredRoute[] {
    return this.registered;
  }
}
