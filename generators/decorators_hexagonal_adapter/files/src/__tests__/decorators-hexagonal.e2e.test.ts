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

  it('GET /api/example/:id returns 200 with valid UUID', async () => {
    const res = await request(app).get('/api/example/33333333-3333-3333-3333-333333333333');
    expect(res.status).toBe(200);
    expect(res.body.id).toBe('33333333-3333-3333-3333-333333333333');
  });

  it('POST /api/example returns 201 when body is valid', async () => {
    const res = await request(app).post('/api/example').send({ name: 'hex-demo' });
    expect(res.status).toBe(201);
    expect(res.body.name).toBe('hex-demo');
  });

  it('GET /docs/openapi.json exposes a valid OpenAPI document', async () => {
    const res = await request(app).get('/docs/openapi.json');
    expect(res.status).toBe(200);
    expect(res.body.paths['/api/example']).toBeDefined();
  });
});
