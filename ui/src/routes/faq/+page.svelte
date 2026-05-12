<script>
</script>

<div class="join join-vertical bg-base-100">
	<div class="collapse-arrow collapse join-item border border-base-300">
		<input type="radio" name="my-accordion-4" checked="checked" />
		<div class="collapse-title font-semibold">How do I create an account?</div>
		<div class="collapse-content text-sm">
			Click the "Sign Up" button in the top right corner and follow the registration process.
		</div>
	</div>
	<div class="collapse-arrow collapse join-item border border-base-300">
		<input type="radio" name="my-accordion-4" />
		<div class="collapse-title font-semibold">I forgot my password. What should I do?</div>
		<div class="collapse-content text-sm">
			Click on "Forgot Password" on the login page and follow the instructions sent to your email.
		</div>
	</div>
	<div class="collapse-arrow collapse join-item border border-base-300">
		<input type="radio" name="my-accordion-4" />
		<div class="collapse-title font-semibold">How do I update my profile information?</div>
		<div class="collapse-content text-sm">
			Go to "My Account" settings and select "Edit Profile" to make changes.
		</div>
	</div>
</div>
<div class="flex flex-wrap items-start gap-3 p-8">
	<!-- Step 1 -->
	<div class="card w-56 border border-base-300 bg-base-100 shadow-sm">
		<div class="card-body gap-2 p-4">
			<p class="text-xs text-base-content/40">step 1</p>
			<h2 class="card-title text-sm">Input Validator</h2>
			<div class="flex flex-wrap gap-1">
				<span class="badge badge-sm badge-error">required</span>
			</div>
			<div class="collapse mt-1 rounded-box border border-base-200">
				<input type="checkbox" />
				<div class="collapse-title min-h-0 px-3 py-2 text-xs font-medium">Details</div>
				<div class="collapse-content px-3 text-xs">
					<p class="opacity-60">Validates incoming payload shape before passing downstream.</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Arrow -->
	<div class="self-center text-xl text-base-content/30">→</div>

	<!-- Step 2 -->
	<div class="card w-56 border border-base-300 bg-base-100 shadow-sm">
		<div class="card-body gap-2 p-4">
			<p class="text-xs text-base-content/40">step 2</p>
			<h2 class="card-title text-sm">Auth Check</h2>
			<div class="flex flex-wrap gap-1">
				<span class="badge badge-sm badge-warning">can-override</span>
			</div>
			<div class="collapse mt-1 rounded-box border border-base-200">
				<input type="checkbox" />
				<div class="collapse-title min-h-0 px-3 py-2 text-xs font-medium">Details</div>
				<div class="collapse-content px-3 text-xs">
					<p class="opacity-60">Verifies token and permissions. Can be bypassed by admin rules.</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Arrow -->
	<div class="self-center text-xl text-base-content/30">→</div>

	<!-- Step 3 -->
	<div class="card w-56 border border-base-300 bg-base-100 shadow-sm">
		<div class="card-body gap-2 p-4">
			<p class="text-xs text-base-content/40">step 3</p>
			<h2 class="card-title text-sm">Rule Evaluator</h2>
			<div class="flex flex-wrap gap-1">
				<span class="badge badge-sm badge-info">async</span>
				<span class="badge badge-sm badge-error">required</span>
			</div>
			<div class="collapse mt-1 rounded-box border border-base-200">
				<input type="checkbox" />
				<div class="collapse-title min-h-0 px-3 py-2 text-xs font-medium">Details</div>
				<div class="collapse-content px-3 text-xs">
					<p class="opacity-60">Runs all matching rules against the context. Async safe.</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Arrow -->
	<div class="self-center text-xl text-base-content/30">→</div>

	<!-- Step 4 -->
	<div class="card w-56 border border-base-300 bg-base-100 shadow-sm">
		<div class="card-body gap-2 p-4">
			<p class="text-xs text-base-content/40">step 4</p>
			<h2 class="card-title text-sm">Output Handler</h2>
			<div class="flex flex-wrap gap-1">
				<span class="badge badge-sm badge-neutral">terminal</span>
			</div>
			<div class="collapse mt-1 rounded-box border border-base-200">
				<input type="checkbox" />
				<div class="collapse-title min-h-0 px-3 py-2 text-xs font-medium">Details</div>
				<div class="collapse-content px-3 text-xs">
					<p class="opacity-60">Formats and emits the final result. No further steps.</p>
				</div>
			</div>
		</div>
	</div>
</div>
