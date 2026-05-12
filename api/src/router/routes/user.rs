use crate::{
    proto::proto::{GetUserRequest, UpdateUserRequest, User},
    route_handlers,
};

route_handlers!(user_svc, {
    get_user (GetUserRequest => User),
    update_user (UpdateUserRequest => User),
});
