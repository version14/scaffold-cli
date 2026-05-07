import type { Express, RequestHandler } from 'express';
import swaggerUi from 'swagger-ui-express';
import { swaggerSpec } from './swagger.config';

export interface MountSwaggerOptions {
  /** URL path that serves the Swagger UI (default: /docs). */
  path?: string;
  /** URL path that serves the raw JSON spec (default: /docs/openapi.json). */
  jsonPath?: string;
}

/**
 * Mount the Swagger UI at /docs and serve the raw spec at /docs/openapi.json.
 *
 * The spec is rebuilt at boot from every `@openapi` JSDoc block found under
 * `src/` (see {@link swaggerSpec}). New routes pick up automatically as long
 * as their JSDoc is well-formed.
 */
export function mountSwagger(app: Express, opts: MountSwaggerOptions = {}): void {
  const uiPath = opts.path ?? '/docs';
  const jsonPath = opts.jsonPath ?? `${uiPath}/openapi.json`;

  const jsonHandler: RequestHandler = (_req, res) => {
    res.json(swaggerSpec);
  };

  app.get(jsonPath, jsonHandler);
  app.use(uiPath, swaggerUi.serve, swaggerUi.setup(swaggerSpec));
}

export { swaggerSpec, swaggerOptions } from './swagger.config';
