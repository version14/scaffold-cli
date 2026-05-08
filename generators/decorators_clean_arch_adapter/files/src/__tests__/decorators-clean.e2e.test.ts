import request from 'supertest';
import { describe, expect, it } from 'vitest';
import app from '../app';

describe('decorator-driven routes (clean architecture, E2E)', () => {
  it('GET /health returns ok', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
  });

  it('GET /api/example/:id returns 400 when id is not a UUID', async () => {
    const res = await request(app).get('/api/example/not-a-uuid');
    expect(res.status).toBe(400);
    expect(res.body.error).toBe('ValidationError');
    expect(res.body.target).toBe('params');
  });

  it('GET /api/example/:id returns 200 with the sample payload when the UUID is valid', async () => {
    const res = await request(app).get('/api/example/11111111-1111-1111-1111-111111111111');
    expect(res.status).toBe(200);
    // Sample controller does not echo the input (see controller comment).
    // It returns the canned synthetic example so the response shape matches
    // the OpenAPI schema regardless of the URL parameter.
    expect(typeof res.body.id).toBe('string');
    expect(res.body.name).toBe('sample');
  });

  it('POST /api/example returns 400 when body is invalid', async () => {
    const res = await request(app).post('/api/example').send({});
    expect(res.status).toBe(400);
    expect(res.body.target).toBe('body');
  });

  it('POST /api/example returns 201 with a synthetic payload when body is valid', async () => {
    const res = await request(app).post('/api/example').send({ name: 'demo' });
    expect(res.status).toBe(201);
    expect(typeof res.body.id).toBe('string');
    expect(res.body.name).toBe('created');
  });

  it('GET /docs/openapi.json exposes a valid OpenAPI document with /api/example paths', async () => {
    const res = await request(app).get('/docs/openapi.json');
    expect(res.status).toBe(200);
    expect(res.body.openapi).toBe('3.0.0');
    expect(res.body.paths['/api/example']).toBeDefined();
    expect(res.body.paths['/api/example/{id}']).toBeDefined();
  });
});
