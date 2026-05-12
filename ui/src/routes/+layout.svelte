<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { resolve } from '$app/paths';
	let { children } = $props();
</script>

<div class="flex h-screen flex-col">
	<div class="navbar bg-base-100 shadow-sm">
		<div class="navbar-start">
			<div class="dropdown">
				<div tabindex="0" role="button" class="btn btn-ghost lg:hidden">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 12h8m-8 6h16"
						/>
					</svg>
				</div>
				<ul
					tabindex="-1"
					class="dropdown-content menu z-1 mt-3 w-52 menu-sm rounded-box bg-base-100 p-2 shadow"
				>
					<li><a href={resolve('/what')}>What is bDNS?</a></li>
					<li>
						How to Setup
						<ul class="p-2">
							<li>Submenu</li>
							<li>Submenu 2</li>
						</ul>
					</li>
					<li><a href={resolve('/faq')}>FAQ</a></li>
				</ul>
			</div>
			<a class="btn text-xl btn-ghost" href={resolve('/')}>bDNS</a>
		</div>
		<div class="navbar-center hidden lg:flex">
			<ul class="menu menu-horizontal px-1">
				<li><a href={resolve('/what')}>What is bDNS?</a></li>
				<li>
					<details>
						<summary>How to Setup</summary>
						<ul class="z-1 w-40 bg-base-100 p-2">
							<li>Submenuu</li>
							<li>Submenu 2</li>
						</ul>
					</details>
				</li>
				<li><a href={resolve('/faq')}>FAQ</a></li>
			</ul>
		</div>
		<div class="navbar-end">
			<a class="btn" href={resolve('/user/login')}>Login/Register</a>
		</div>
	</div>
	<main class="grow overflow-auto">
		{@render children()}
	</main>
</div>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
