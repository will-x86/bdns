use crate::{
    proto::proto::{LoginResponse, SignUpResponse},
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
    pub async fn sign_up(&self, timezone: String) -> Result<SignUpResponse, AppError> {
        use super::proto::SignUpRequest;
        let request = SignUpRequest { timezone };
        let response = self.client.clone().sign_up(request).await?;
        let inner = response.into_inner();
        Ok(inner)
    }

    pub async fn login(&self, user_id: String) -> Result<LoginResponse, AppError> {
        use super::proto::LoginRequest;
        let request = LoginRequest { user_id };
        let response = self.client.clone().login(request).await?;
        let inner = response.into_inner();
        Ok(inner)
    }
}
