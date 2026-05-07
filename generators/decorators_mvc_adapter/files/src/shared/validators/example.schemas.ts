import { z } from '../openapi/registry';

export const exampleParamsSchema = z.object({
  id: z.string().uuid(),
});

export const exampleCreateSchema = z.object({
  name: z.string().min(1).max(120),
  description: z.string().max(500).optional(),
});

export const exampleResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  description: z.string().nullable(),
});

export type ExampleParams = z.infer<typeof exampleParamsSchema>;
export type ExampleCreate = z.infer<typeof exampleCreateSchema>;
export type ExampleResponse = z.infer<typeof exampleResponseSchema>;
