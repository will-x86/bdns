use crate::{
    proto::proto::{
        AddPermanentRequest, AddTemporaryRequest, ListPermanentRequest, ListTemporaryRequest,
        RemovePermanentRequest, RemoveTemporaryRequest, WhitelistDomain, WhitelistDomainTemp,
        WhitelistDomains, WhitelistDomainsTemp,
    },
    route_handlers,
};

route_handlers!(whitelist_svc, {
    list_permanent (ListPermanentRequest => WhitelistDomains),
    add_permanent (AddPermanentRequest => WhitelistDomain),
    remove_permanent (RemovePermanentRequest => ()),
    list_temporary (ListTemporaryRequest => WhitelistDomainsTemp),
    add_temporary (AddTemporaryRequest => WhitelistDomainTemp),
    remove_temporary (RemoveTemporaryRequest => ()),
});
