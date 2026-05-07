import 'reflect-metadata';
import express from 'express';
import cors from 'cors';
import {
  DecoratorRouter,
  ExpressRouterAdapter,
} from './shared/decorators';
import { buildOpenApiSpec, createRegistry, mountSwagger } from './shared/openapi';
import { ExampleController } from './modules/example/application/controllers/example.controller';

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

const decoratorRouter = new DecoratorRouter(new ExpressRouterAdapter())
  .registerController(new ExampleController());

app.use(decoratorRouter.build());

const spec = buildOpenApiSpec({
  info: { title: 'API', version: '1.0.0' },
  servers: [{ url: '/' }],
  routes: decoratorRouter.routes(),
  registry: createRegistry(),
});
mountSwagger(app, spec);

export default app;
