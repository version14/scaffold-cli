import express from 'express';
import request from 'supertest';
import { describe, expect, it } from 'vitest';
import { z } from 'zod';
import {
  ApiResponse,
  Auth,
  Body,
  Controller,
  DecoratorRouter,
  ExpressRouterAdapter,
  Get,
  Params,
  Post,
  Query,
  RequiredHeaders,
} from '..';

function buildApp(controller: object, options: { authMiddleware?: express.RequestHandler } = {}) {
  const adapter = new ExpressRouterAdapter(options);
  const router = new DecoratorRouter(adapter).registerController(controller).build();
  const app = express();
  app.use(express.json());
  app.use(router);
  return { app, adapter };
}

describe('decorators (unit)', () => {
  it('registers a GET route declared via @Get', async () => {
    @Controller({ tag: 'health', prefix: '/health' })
    class HealthController {
      @Get('/')
      @ApiResponse(200, 'ok')
      ping(_req: express.Request, res: express.Response) {
        res.json({ status: 'ok' });
      }
    }

    const { app } = buildApp(new HealthController());
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: 'ok' });
  });

  it('joins controller prefix and route path correctly', async () => {
    @Controller({ tag: 'users', prefix: '/users' })
    class UsersController {
      @Get(':id')
      get(req: express.Request, res: express.Response) {
        res.json({ id: req.params.id });
      }
    }

    const { app } = buildApp(new UsersController());
    const res = await request(app).get('/users/42');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ id: '42' });
  });

  it('returns 400 with structured issues when @Body validation fails', async () => {
    const schema = z.object({ email: z.string().email(), age: z.number().int().min(0) });

    @Controller({ tag: 'signup', prefix: '/signup' })
    class SignupController {
      @Post('/')
      @Body(schema)
      handle(_req: express.Request, res: express.Response) {
        res.status(201).json({ created: true });
      }
    }

    const { app } = buildApp(new SignupController());
    const res = await request(app).post('/signup').send({ email: 'nope', age: -1 });
    expect(res.status).toBe(400);
    expect(res.body.error).toBe('ValidationError');
    expect(res.body.target).toBe('body');
    expect(res.body.issues.length).toBeGreaterThan(0);
  });

  it('passes through with parsed body when @Body validation succeeds', async () => {
    const schema = z.object({ name: z.string() });

    @Controller({ tag: 'echo', prefix: '/echo' })
    class EchoController {
      @Post('/')
      @Body(schema)
      handle(req: express.Request, res: express.Response) {
        res.json(req.body);
      }
    }

    const { app } = buildApp(new EchoController());
    const res = await request(app).post('/echo').send({ name: 'mathieu' });
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ name: 'mathieu' });
  });

  it('coerces query params via @Query schema', async () => {
    const schema = z.object({ page: z.coerce.number().int().min(1) });

    @Controller({ tag: 'list', prefix: '/list' })
    class ListController {
      @Get('/')
      @Query(schema)
      handle(req: express.Request, res: express.Response) {
        res.json(req.query);
      }
    }

    const { app } = buildApp(new ListController());
    const res = await request(app).get('/list?page=3');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ page: 3 });
  });

  it('rejects invalid params via @Params schema', async () => {
    const schema = z.object({ id: z.string().uuid() });

    @Controller({ tag: 'items', prefix: '/items' })
    class ItemsController {
      @Get(':id')
      @Params(schema)
      handle(_req: express.Request, res: express.Response) {
        res.json({ ok: true });
      }
    }

    const { app } = buildApp(new ItemsController());
    const res = await request(app).get('/items/not-a-uuid');
    expect(res.status).toBe(400);
    expect(res.body.target).toBe('params');
  });

  it('blocks @Auth-protected route with 401 when no auth middleware accepts the request', async () => {
    const denyAll: express.RequestHandler = (_req, res) => {
      res.status(401).json({ error: 'Unauthorized' });
    };

    @Controller({ tag: 'private', prefix: '/private' })
    class PrivateController {
      @Get('/')
      @Auth()
      handle(_req: express.Request, res: express.Response) {
        res.json({ ok: true });
      }
    }

    const { app } = buildApp(new PrivateController(), { authMiddleware: denyAll });
    const res = await request(app).get('/private');
    expect(res.status).toBe(401);
  });

  it('lets @Auth-protected route through when auth middleware accepts the request', async () => {
    const acceptAll: express.RequestHandler = (_req, _res, next) => next();

    @Controller({ tag: 'private', prefix: '/private' })
    class PrivateController {
      @Get('/')
      @Auth()
      handle(_req: express.Request, res: express.Response) {
        res.json({ ok: true });
      }
    }

    const { app } = buildApp(new PrivateController(), { authMiddleware: acceptAll });
    const res = await request(app).get('/private');
    expect(res.status).toBe(200);
  });

  it('returns 400 when a @RequiredHeaders header is missing', async () => {
    @Controller({ tag: 'gated', prefix: '/gated' })
    class GatedController {
      @Get('/')
      @RequiredHeaders(['x-tenant'])
      handle(_req: express.Request, res: express.Response) {
        res.json({ ok: true });
      }
    }

    const { app } = buildApp(new GatedController());
    const res = await request(app).get('/gated');
    expect(res.status).toBe(400);
    expect(res.body.error).toBe('MissingHeaders');
  });

  it('exposes registered routes for downstream consumers (e.g. OpenAPI)', () => {
    const schema = z.object({ name: z.string() });

    @Controller({ tag: 'meta', prefix: '/meta' })
    class MetaController {
      @Post('/')
      @Body(schema)
      @ApiResponse(201, 'created', schema)
      handle(_req: express.Request, res: express.Response) {
        res.status(201).json({});
      }
    }

    const adapter = new ExpressRouterAdapter();
    const router = new DecoratorRouter(adapter).registerController(new MetaController());
    const registered = router.routes();

    expect(registered).toHaveLength(1);
    expect(registered[0]?.fullPath).toBe('/meta');
    expect(registered[0]?.route.method).toBe('post');
    expect(registered[0]?.validation.body).toBe(schema);
    expect(registered[0]?.responses[0]?.status).toBe(201);
  });
});
