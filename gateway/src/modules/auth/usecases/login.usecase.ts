import { Inject, InternalServerErrorException } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { AuthServiceClient } from 'protos/ts/auth/auth';
import { LoginRequest, LoginResponse } from 'protos/ts/gateway/gateway';
import { firstValueFrom } from 'rxjs';

export class LoginUseCase {
  private authService: AuthServiceClient;

  constructor(@Inject('AUTH_SERVICE') private readonly client: ClientGrpc) {
    this.authService = this.client.getService('AuthService');
  }

  async execute(req: LoginRequest): Promise<LoginResponse> {
    const observableResponse = this.authService.login(req);
    const response = await firstValueFrom(observableResponse).catch((error) => {
      throw new RpcException(error as object);
    });

    const tokens = response.tokens;
    if (!tokens) {
      throw new RpcException(
        new InternalServerErrorException('Could not login'),
      );
    }

    return { tokens };
  }
}
