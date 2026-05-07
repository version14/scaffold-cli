import type { NextFunction, Request, RequestHandler, Response } from 'express';
import type { ZodSchema } from 'zod';

export type ValidationTarget = 'body' | 'params' | 'query';

/**
 * Returns an Express middleware that validates `req[target]` against the
 * supplied Zod schema. On success, `req[target]` is replaced by the parsed
 * (and coerced) value so downstream handlers see a typed payload.
 *
 * On failure, responds with 400 and a structured error payload listing every
 * issue Zod surfaced. The handler is never invoked.
 */
export function validateRequest(schema: ZodSchema, target: ValidationTarget): RequestHandler {
  return (req: Request, res: Response, next: NextFunction): void => {
    const result = schema.safeParse(req[target]);
    if (!result.success) {
      res.status(400).json({
        error: 'ValidationError',
        target,
        issues: result.error.issues.map((issue) => ({
          path: issue.path,
          message: issue.message,
          code: issue.code,
        })),
      });
      return;
    }
    (req as unknown as Record<string, unknown>)[target] = result.data;
    next();
  };
}
