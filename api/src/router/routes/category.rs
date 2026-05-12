use crate::{
    proto::proto::{
        BlockCategoryRequest, CategoryBlock, CategoryList, ListBlockedRequest,
        UnblockCategoryRequest,
    },
    route_handlers,
};

route_handlers!(category_svc, {
    list_blocked (ListBlockedRequest => CategoryList),
    block_category (BlockCategoryRequest => CategoryBlock),
    unblock_category (UnblockCategoryRequest => ()),
});
