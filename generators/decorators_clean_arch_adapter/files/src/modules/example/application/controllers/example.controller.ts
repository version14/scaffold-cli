import type { Request, Response } from 'express';
import {
  ApiResponse,
  Body,
  Controller,
  Get,
  Params,
  Post,
} from '../../../../shared/decorators';
import {
  exampleCreateSchema,
  exampleParamsSchema,
  exampleResponseSchema,
  type ExampleCreate,
  type ExampleParams,
} from '../validators/example.schemas';

/**
 * Example controller — kept minimal on purpose. Replace this with your real
 * use cases (typically injected through the constructor) and remove this file
 * once you no longer need the reference.
 *
 * Design note: the synthetic response payloads below intentionally do **not**
 * mirror the request input. Wire the validated DTOs (`params`, `body`) to your
 * repositories / use cases and return the persisted entity instead.
 */
@Controller({ tag: 'example', prefix: '/api/example', description: 'Sample resource demonstrating decorator usage' })
export class ExampleController {
  @Get(':id')
  @Params(exampleParamsSchema)
  @ApiResponse(200, 'Example fetched', exampleResponseSchema)
  @ApiResponse(404, 'Example not found')
  get(req: Request, res: Response): void {
    // params is validated by @Params and typed as ExampleParams.
    // Use it to call your repository: `repo.findById(params.id)`.
    const params = req.params as unknown as ExampleParams;
    void params;
    res.json({
      id: '11111111-1111-1111-1111-111111111111',
      name: 'sample',
      description: null,
    });
  }

  @Post('/')
  @Body(exampleCreateSchema)
  @ApiResponse(201, 'Example created', exampleResponseSchema)
  create(req: Request, res: Response): void {
    // body is validated by @Body and typed as ExampleCreate.
    // Forward it to a CreateExampleUseCase and return the persisted entity.
    const body = req.body as ExampleCreate;
    void body;
    res.status(201).json({
      id: '00000000-0000-0000-0000-000000000000',
      name: 'created',
      description: null,
    });
  }
}
