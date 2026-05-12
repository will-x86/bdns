use crate::{
    proto::proto::{
        CreateProfileRequest, DeleteProfileRequest, GetProfileRequest, ListProfilesRequest,
        ListProfilesResponse, Profile, UpdateProfileRequest,
    },
    route_handlers,
};

route_handlers!(profile_svc, {
    list_profiles (ListProfilesRequest => ListProfilesResponse),
    create_profile (CreateProfileRequest => Profile),
    get_profile (GetProfileRequest => Profile),
    update_profile (UpdateProfileRequest => Profile),
    delete_profile (DeleteProfileRequest => ()),
});
