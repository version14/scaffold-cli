import type { Express, RequestHandler } from 'express';
import swaggerUi from 'swagger-ui-express';

export interface MountSwaggerOptions {
  /** URL path that serves the Swagger UI (default: /docs). */
  path?: string;
  /** URL path that serves the raw JSON spec (default: /docs/openapi.json). */
  jsonPath?: string;
}

/**
 * Serve a Swagger UI for the given OpenAPI spec at /docs (configurable). The
 * raw JSON document is also served so external clients (Postman, codegen) can
 * consume it programmatically.
 */
export function mountSwagger(app: Express, spec: Record<string, unknown>, opts: MountSwaggerOptions = {}): void {
  const path = opts.path ?? '/docs';
  const jsonPath = opts.jsonPath ?? `${path}/openapi.json`;

  const jsonHandler: RequestHandler = (_req, res) => {
    res.json(spec);
  };

  app.get(jsonPath, jsonHandler);
  app.use(path, swaggerUi.serve, swaggerUi.setup(spec));
}
