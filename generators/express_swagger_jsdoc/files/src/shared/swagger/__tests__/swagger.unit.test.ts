import express from 'express';
import request from 'supertest';
import { describe, expect, it } from 'vitest';
import { mountSwagger } from '..';

describe('JSDoc-based Swagger', () => {
  it('serves /docs/openapi.json with a valid v3 document', async () => {
    const app = express();
    mountSwagger(app);

    const res = await request(app).get('/docs/openapi.json');
    expect(res.status).toBe(200);
    expect(res.body.openapi).toBe('3.0.0');
    expect(res.body.info?.title).toBe('API');
    // The /health endpoint declared in src/app.ts should be picked up by the scanner.
    expect(res.body.paths?.['/health']).toBeDefined();
  });

  it('honours custom mount path', async () => {
    const app = express();
    mountSwagger(app, { path: '/api-docs' });

    const json = await request(app).get('/api-docs/openapi.json');
    expect(json.status).toBe(200);
  });
});
