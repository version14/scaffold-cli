import request from 'supertest';
import { describe, expect, it } from 'vitest';
import app from '../app';

describe('decorator-driven routes (hexagonal, E2E)', () => {
  it('GET /health returns ok', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
  });

  it('GET /api/example/:id returns 400 when id is not a UUID', async () => {
    const res = await request(app).get('/api/example/not-a-uuid');
    expect(res.status).toBe(400);
  });

  it('GET /api/example/:id returns 200 with the sample payload when the UUID is valid', async () => {
    const res = await request(app).get('/api/example/33333333-3333-3333-3333-333333333333');
    expect(res.status).toBe(200);
    // Sample controller does not echo the input (see controller comment).
    expect(typeof res.body.id).toBe('string');
    expect(res.body.name).toBe('sample');
  });

  it('POST /api/example returns 201 with a synthetic payload when body is valid', async () => {
    const res = await request(app).post('/api/example').send({ name: 'hex-demo' });
    expect(res.status).toBe(201);
    expect(typeof res.body.id).toBe('string');
    expect(res.body.name).toBe('created');
  });

  it('GET /docs/openapi.json exposes a valid OpenAPI document', async () => {
    const res = await request(app).get('/docs/openapi.json');
    expect(res.status).toBe(200);
    expect(res.body.paths['/api/example']).toBeDefined();
  });
});
