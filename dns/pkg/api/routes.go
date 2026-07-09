package api

import "github.com/gofiber/fiber/v3"

// routes registers every endpoint. Anything after v1.Use(a.auth) is protected.
func (a *API) routes(app *fiber.App) {
	v1 := app.Group("/api/v1")

	// Public.
	v1.Get("/health", a.health)
	v1.Post("/users", a.createUser) // bootstrap: mints a user + token

	// Protected.
	v1.Use(a.auth)

	// Current user.
	v1.Get("/me", a.getMe)
	v1.Patch("/me", a.updateMe)
	v1.Delete("/me", a.deleteMe)
	v1.Post("/me/token", a.regenerateToken)

	// Friends.
	v1.Get("/friends", a.listFriends)
	v1.Post("/friends", a.addFriend)
	v1.Delete("/friends/:friendID", a.deleteFriend)

	// Blocklist categories (for UI dropdowns).
	v1.Get("/categories", a.listCategories)

	// Profiles.
	v1.Get("/profiles", a.listProfiles)
	v1.Post("/profiles", a.createProfile)
	v1.Get("/profiles/:pid", a.getProfile)
	v1.Patch("/profiles/:pid", a.updateProfile)
	v1.Delete("/profiles/:pid", a.deleteProfile)

	// Permanent whitelist.
	v1.Get("/profiles/:pid/whitelist/permanent", a.listPermanentWhitelist)
	v1.Post("/profiles/:pid/whitelist/permanent", a.addPermanentWhitelist)
	v1.Delete("/profiles/:pid/whitelist/permanent/:domain", a.deletePermanentWhitelist)

	// Temporary whitelist.
	v1.Get("/profiles/:pid/whitelist/temporary", a.listTemporaryWhitelist)
	v1.Post("/profiles/:pid/whitelist/temporary", a.addTemporaryWhitelist)
	v1.Delete("/profiles/:pid/whitelist/temporary/:domain", a.deleteTemporaryWhitelist)

	// Category blocks.
	v1.Get("/profiles/:pid/category-blocks", a.listCategoryBlocks)
	v1.Post("/profiles/:pid/category-blocks", a.addCategoryBlock)
	v1.Delete("/profiles/:pid/category-blocks/:category", a.deleteCategoryBlock)

	// Time blocks.
	v1.Get("/profiles/:pid/time-blocks", a.listTimeBlocks)
	v1.Post("/profiles/:pid/time-blocks", a.addTimeBlock)
	v1.Delete("/profiles/:pid/time-blocks", a.deleteTimeBlock)

	// Pools.
	v1.Get("/pools", a.listPools)
	v1.Post("/pools", a.createPool)
	v1.Get("/pools/:poolID", a.getPool)
	v1.Patch("/pools/:poolID", a.updatePool)
	v1.Delete("/pools/:poolID", a.deletePool)
	v1.Get("/pools/:poolID/limits", a.getPoolLimits)

	// Pool members.
	v1.Get("/pools/:poolID/members", a.listPoolMembers)
	v1.Post("/pools/:poolID/members", a.addPoolMember)
	v1.Delete("/pools/:poolID/members/:profileID", a.deletePoolMember)

	// Pool category blocks.
	v1.Get("/pools/:poolID/category-blocks", a.listPoolCategoryBlocks)
	v1.Post("/pools/:poolID/category-blocks", a.addPoolCategoryBlock)
	v1.Delete("/pools/:poolID/category-blocks/:category", a.deletePoolCategoryBlock)
}

func (a *API) health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
