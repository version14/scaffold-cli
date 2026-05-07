import express from 'express';
import request from 'supertest';
import { describe, expect, it } from 'vitest';
import {
  ApiResponse,
  Body,
  Controller,
  DecoratorRouter,
  ExpressRouterAdapter,
  Get,
  Post,
  Auth,
} from '../../decorators';
import { createRegistry } from '../registry';
import { buildOpenApiSpec } from '../spec-generator';
import { mountSwagger } from '../swagger';
import { z } from 'zod';

describe('OpenAPI spec generation', () => {
  it('produces a v3 document with the expected paths and tags', () => {
    const userSchema = z.object({ id: z.string(), email: z.string().email() });

    @Controller({ tag: 'users', prefix: '/users', description: 'User management' })
    class UsersController {
      @Get(':id')
      @ApiResponse(200, 'user found', userSchema)
      get(_req: express.Request, res: express.Response) {
        res.json({});
      }

      @Post('/')
      @Body(userSchema)
      @ApiResponse(201, 'user created', userSchema)
      create(_req: express.Request, res: express.Response) {
        res.json({});
      }
    }

    const router = new DecoratorRouter(new ExpressRouterAdapter()).registerController(new UsersController());
    const spec = buildOpenApiSpec({
      info: { title: 'Test API', version: '1.0.0' },
      servers: [{ url: '/api' }],
      routes: router.routes(),
      registry: createRegistry(),
    });

    expect(spec.openapi).toBe('3.0.0');
    const paths = spec.paths as Record<string, Record<string, unknown>>;
    expect(paths['/users/{id}']).toBeDefined();
    expect(paths['/users']).toBeDefined();

    const getOp = paths['/users/{id}']?.get as { tags?: string[]; responses: Record<string, unknown> };
    expect(getOp?.tags).toEqual(['users']);
    expect(getOp?.responses['200']).toBeDefined();

    const postOp = paths['/users']?.post as { requestBody: unknown };
    expect(postOp?.requestBody).toBeDefined();
  });

  it('marks @Auth-decorated routes with BearerAuth security', () => {
    @Controller({ tag: 'me', prefix: '/me' })
    class MeController {
      @Get('/')
      @Auth()
      @ApiResponse(200, 'profile')
      handle(_req: express.Request, res: express.Response) {
        res.json({});
      }
    }

    const router = new DecoratorRouter(new ExpressRouterAdapter()).registerController(new MeController());
    const spec = buildOpenApiSpec({
      info: { title: 'API', version: '1.0.0' },
      routes: router.routes(),
      registry: createRegistry(),
    });

    const op = (spec.paths as Record<string, { get: { security: Array<Record<string, unknown>> } }>)['/me']?.get;
    expect(op?.security).toEqual([{ BearerAuth: [] }]);
  });

  it('falls back to a default 200 response when none is declared', () => {
    @Controller({ tag: 'noop', prefix: '/noop' })
    class NoopController {
      @Get('/')
      handle(_req: express.Request, res: express.Response) {
        res.send();
      }
    }

    const router = new DecoratorRouter(new ExpressRouterAdapter()).registerController(new NoopController());
    const spec = buildOpenApiSpec({
      info: { title: 'API', version: '1.0.0' },
      routes: router.routes(),
      registry: createRegistry(),
    });

    const op = (spec.paths as Record<string, { get: { responses: Record<string, { description: string }> } }>)['/noop']
      ?.get;
    expect(op?.responses['200']?.description).toBe('Successful response');
  });

  it('serves /docs/openapi.json with the spec', async () => {
    @Controller({ tag: 'health', prefix: '/health' })
    class HealthController {
      @Get('/')
      @ApiResponse(200, 'ok')
      handle(_req: express.Request, res: express.Response) {
        res.json({});
      }
    }

    const router = new DecoratorRouter(new ExpressRouterAdapter()).registerController(new HealthController());
    const spec = buildOpenApiSpec({
      info: { title: 'Mounted API', version: '0.0.1' },
      routes: router.routes(),
      registry: createRegistry(),
    });

    const app = express();
    mountSwagger(app, spec);

    const res = await request(app).get('/docs/openapi.json');
    expect(res.status).toBe(200);
    expect(res.body.info.title).toBe('Mounted API');
    expect(res.body.paths['/health']).toBeDefined();
  });
});
