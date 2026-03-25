namespace go user

struct UserInfo {
  1: i64 user_id
  2: string email
  3: string nickname
}

struct CreateUserRequest {
  1: string email
  2: string nickname
  3: string password
}

struct CreateUserResponse {
  1: UserInfo user
}

struct GetUserRequest {
  1: i64 user_id
}

struct GetUserResponse {
  1: UserInfo user
}

struct VerifyCredentialRequest {
  1: string email
  2: string password
}

struct VerifyCredentialResponse {
  1: bool ok
  2: UserInfo user
  3: string reason
}

struct CheckPasswordRequest {
  1: i64 user_id
  2: string password
}

struct CheckPasswordResponse {
  1: bool ok
}

service UserService {
  CreateUserResponse CreateUser(1: CreateUserRequest req)
  GetUserResponse GetUser(1: GetUserRequest req)
  VerifyCredentialResponse VerifyCredential(1: VerifyCredentialRequest req)
  CheckPasswordResponse CheckPassword(1: CheckPasswordRequest req)
}
