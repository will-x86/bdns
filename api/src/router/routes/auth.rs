use crate::{
    proto::proto::{LoginRequest, LoginResponse, SignUpRequest, SignUpResponse},
    route_handlers,
};

route_handlers!(auth_svc, {
    sign_up (SignUpRequest => SignUpResponse),
    login (LoginRequest => LoginResponse),
});
