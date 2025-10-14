import {
  BadRequestException,
  Inject,
  Injectable,
  InternalServerErrorException,
} from '@nestjs/common';
import { GetUserResponse, UserServiceClient } from 'protos/ts/user/user';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import {
  GetUserProfileRequest,
  GetUserProfileResponse,
} from 'protos/ts/gateway/gateway';
import { firstValueFrom, Observable } from 'rxjs';

@Injectable()
export class GetUserProfileUseCase {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService = this.client.getService<UserServiceClient>('UserService');
  }

  async execute({
    id,
    username,
  }: GetUserProfileRequest): Promise<GetUserProfileResponse> {
    let observableResponse: Observable<GetUserResponse>;

    if (id) {
      observableResponse = this.userService.getUserById({ id });
    } else if (username) {
      observableResponse = this.userService.getUserByUsername({ username });
    } else {
      throw new RpcException(
        new BadRequestException('Id or username is required'),
      );
    }

    const serviceResponse = await firstValueFrom(observableResponse).catch(
      (error) => {
        throw new RpcException(error as object);
      },
    );

    const user = serviceResponse.user;
    if (!user) {
      throw new RpcException(
        new InternalServerErrorException('Something went wrong'),
      );
    }

    return {
      id: user.id,
      name: user.name,
      username: user.username,
    };
  }
}
