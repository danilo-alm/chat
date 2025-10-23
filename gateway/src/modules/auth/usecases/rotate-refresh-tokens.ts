import {
  Inject,
  Injectable,
  InternalServerErrorException,
} from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import {
  AuthServiceClient,
  RotateRefreshTokenRequest,
  RotateRefreshTokenResponse,
} from 'protos/ts/auth/auth';
import { firstValueFrom } from 'rxjs';

@Injectable()
export class RotateRefreshTokensUseCase {
  private authService: AuthServiceClient;

  constructor(@Inject('AUTH_SERVICE') private readonly client: ClientGrpc) {
    this.authService = this.client.getService('AuthService');
  }

  async execute(
    req: RotateRefreshTokenRequest,
  ): Promise<RotateRefreshTokenResponse> {
    const observableResponse = this.authService.rotateRefreshToken(req);
    const response = await firstValueFrom(observableResponse).catch((error) => {
      throw new RpcException(error as object);
    });

    const tokens = response.tokens;
    if (!tokens) {
      throw new RpcException(
        new InternalServerErrorException('Could not rotate tokens'),
      );
    }

    return { tokens };
  }
}
