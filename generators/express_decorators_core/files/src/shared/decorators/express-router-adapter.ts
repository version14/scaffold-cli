import { type RequestHandler, Router } from 'express';
import { validateRequest } from '../middlewares/validate-request';
import type { RouteRegistration, RouterAdapter } from './router-adapter';

export interface ExpressAdapterOptions {
  authMiddleware?: RequestHandler;
}

function requiredHeadersMiddleware(headers: string[]): RequestHandler {
  return (req, res, next) => {
    const missing = headers.filter((h) => !req.headers[h.toLowerCase()]);
    if (missing.length > 0) {
      res.status(400).json({ error: 'MissingHeaders', headers: missing });
      return;
    }
    next();
  };
}

/**
 * Wrap a possibly-async handler so any thrown error or rejected promise is
 * forwarded to Express' error pipeline via next(err) instead of becoming an
 * unhandled rejection. Equivalent to the express-async-errors pattern but
 * scoped to decorator-registered routes.
 */
function asyncSafe(handler: RouteRegistration['handler']): RequestHandler {
  return (req, res, next) => {
    try {
      const result = (handler as (...args: unknown[]) => unknown)(req, res, next);
      if (result instanceof Promise) {
        result.catch(next);
      }
    } catch (err) {
      next(err);
    }
  };
}

/**
 * Express implementation of RouterAdapter. Translates each RouteRegistration
 * into an ordered middleware chain:
 *
 *   [requiredHeaders?] → [auth?] → [validate(params)?] → [validate(query)?] →
 *   [validate(body)?] → handler
 */
export class ExpressRouterAdapter implements RouterAdapter<Router> {
  private readonly router: Router;
  private readonly authMiddleware?: RequestHandler;

  constructor(options: ExpressAdapterOptions = {}) {
    this.router = Router();
    this.authMiddleware = options.authMiddleware;
  }

  register(registration: RouteRegistration): void {
    const middlewares: RequestHandler[] = [];

    if (registration.requiredHeaders.length > 0) {
      middlewares.push(requiredHeadersMiddleware(registration.requiredHeaders));
    }

    if (registration.isProtected && this.authMiddleware) {
      middlewares.push(this.authMiddleware);
    }

    if (registration.validation.params) {
      middlewares.push(validateRequest(registration.validation.params, 'params'));
    }
    if (registration.validation.query) {
      middlewares.push(validateRequest(registration.validation.query, 'query'));
    }
    if (registration.validation.body) {
      middlewares.push(validateRequest(registration.validation.body, 'body'));
    }

    middlewares.push(asyncSafe(registration.handler));

    this.router[registration.method](registration.path, ...middlewares);
  }

  build(): Router {
    return this.router;
  }
}
