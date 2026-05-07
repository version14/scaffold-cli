import path from 'node:path';
import swaggerJSDoc, { type Options } from 'swagger-jsdoc';

const projectRoot = path.resolve(process.cwd());

/**
 * swagger-jsdoc options. Scans every `.ts` and `.js` file under `src/` for
 * `@openapi` JSDoc blocks and assembles them into a single OpenAPI v3
 * document. Update `info.title`/`info.version` to match your project.
 */
export const swaggerOptions: Options = {
  definition: {
    openapi: '3.0.0',
    info: {
      title: 'API',
      version: '1.0.0',
      description: 'API documentation generated from JSDoc @openapi comments.',
    },
    components: {
      securitySchemes: {
        BearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT',
        },
      },
    },
  },
  apis: [
    path.join(projectRoot, 'src/**/*.ts'),
    path.join(projectRoot, 'src/**/*.js'),
  ],
};

export const swaggerSpec = swaggerJSDoc(swaggerOptions);
