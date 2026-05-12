use crate::{
    proto::proto::{LoginRequest, LoginResponse, SignUpRequest, SignUpResponse},
    router::AppError,
};

use super::proto::auth_client::AuthClient;
use tonic::transport::Channel;

#[derive(Clone)]
pub struct AuthSvc {
    client: AuthClient<Channel>,
}

impl AuthSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = AuthClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn sign_up(&self, req: SignUpRequest) -> Result<SignUpResponse, AppError> {
        let response = self.client.clone().sign_up(req).await?;
        Ok(response.into_inner())
    }

    pub async fn login(&self, req: LoginRequest) -> Result<LoginResponse, AppError> {
        let response = self.client.clone().login(req).await?;
        Ok(response.into_inner())
    }
}
