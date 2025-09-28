import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { USER_INJECTION_TOKEN } from '../user.module';
import { ClientGrpc } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { CreateUserRequest, UserServiceClient } from 'protos/ts/user/user';
import {
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class RegisterUserUseCase implements OnModuleInit {
  private userServiceClient: UserServiceClient;

  constructor(
    @Inject(USER_INJECTION_TOKEN) private readonly client: ClientGrpc,
  ) {}

  onModuleInit() {
    this.userServiceClient =
      this.client.getService<UserServiceClient>('UserService');
  }

  async execute(request: RegisterUserRequest): Promise<RegisterUserResponse> {
    const serviceRequest: CreateUserRequest = {
      name: request.name,
      username: request.username,
    };

    const observableResponse =
      this.userServiceClient.createUser(serviceRequest);
    const serviceResponse = await firstValueFrom(observableResponse);

    return {
      id: serviceResponse.id,
    };
  }
}
