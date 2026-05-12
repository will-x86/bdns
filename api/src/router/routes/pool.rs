use crate::{
    proto::proto::{
        BlockPoolCategoryRequest, CreatePoolRequest, CreditsResponse, DeletePoolRequest,
        FriendPool, GetCreditsRequest, GetPoolRequest, JoinPoolRequest, LeavePoolRequest,
        ListMembersRequest, ListPoolBlocksRequest, ListPoolsRequest, PoolBlockList, PoolList,
        PoolMemberList, UnblockPoolCategoryRequest,
    },
    route_handlers,
};

route_handlers!(pool_svc, {
    list_pools (ListPoolsRequest => PoolList),
    create_pool (CreatePoolRequest => FriendPool),
    get_pool (GetPoolRequest => FriendPool),
    delete_pool (DeletePoolRequest => ()),
    join_pool (JoinPoolRequest => ()),
    leave_pool (LeavePoolRequest => ()),
    list_members (ListMembersRequest => PoolMemberList),
    list_blocks (ListPoolBlocksRequest => PoolBlockList),
    block_category (BlockPoolCategoryRequest => ()),
    unblock_category (UnblockPoolCategoryRequest => ()),
    get_credits (GetCreditsRequest => CreditsResponse),
});
