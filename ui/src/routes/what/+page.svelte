<script lang="ts">
	import { Svelvet, Node, Anchor } from 'svelvet';

	type NodeDef = {
		id: string;
		label: string;
		sub?: string;
		x: number;
		y: number;
		w: number;
		h: number;
		bgClass: string;
		textClass: string;
		east?: string;
		south?: string;
	};

	const nodes: NodeDef[] = [
		{
			id: 'request',
			label: 'DNS request',
			x: 320,
			y: 20,
			w: 180,
			h: 44,
			bgClass: 'bg-neutral',
			textClass: 'text-neutral-content',
			south: 'perm-whitelist'
		},
		{
			id: 'perm-whitelist',
			label: 'Permanent whitelist',
			sub: 'Always whitelisted for this profile?',
			x: 240,
			y: 110,
			w: 270,
			h: 60,
			bgClass: 'bg-info',
			textClass: 'text-info-content',
			east: 'allow-1',
			south: 'temp-whitelist'
		},
		{
			id: 'allow-1',
			label: 'Allow',
			x: 570,
			y: 122,
			w: 90,
			h: 36,
			bgClass: 'bg-success',
			textClass: 'text-success-content'
		},
		{
			id: 'temp-whitelist',
			label: 'Temporary whitelist',
			sub: 'Whitelisted for this time window?',
			x: 240,
			y: 216,
			w: 270,
			h: 60,
			bgClass: 'bg-info',
			textClass: 'text-info-content',
			east: 'allow-2',
			south: 'cat-block'
		},
		{
			id: 'allow-2',
			label: 'Allow',
			x: 570,
			y: 228,
			w: 90,
			h: 36,
			bgClass: 'bg-success',
			textClass: 'text-success-content'
		},
		{
			id: 'cat-block',
			label: 'Category block',
			sub: 'Domain category blocked for profile?',
			x: 240,
			y: 322,
			w: 270,
			h: 60,
			bgClass: 'bg-secondary',
			textClass: 'text-secondary-content',
			east: 'block-1',
			south: 'time-block'
		},
		{
			id: 'block-1',
			label: 'Block',
			x: 570,
			y: 334,
			w: 90,
			h: 36,
			bgClass: 'bg-error',
			textClass: 'text-error-content'
		},
		{
			id: 'time-block',
			label: 'Time block',
			sub: 'Active schedule block for category?',
			x: 240,
			y: 428,
			w: 270,
			h: 60,
			bgClass: 'bg-secondary',
			textClass: 'text-secondary-content',
			east: 'block-2',
			south: 'shared-pool'
		},
		{
			id: 'block-2',
			label: 'Block',
			x: 570,
			y: 440,
			w: 90,
			h: 36,
			bgClass: 'bg-error',
			textClass: 'text-error-content'
		},
		{
			id: 'shared-pool',
			label: 'Shared pool',
			sub: 'Pool=shared, category blocked, no quota?',
			x: 240,
			y: 534,
			w: 270,
			h: 64,
			bgClass: 'bg-warning',
			textClass: 'text-warning-content',
			east: 'block-3',
			south: 'borrow-pool'
		},
		{
			id: 'block-3',
			label: 'Block',
			x: 570,
			y: 550,
			w: 90,
			h: 36,
			bgClass: 'bg-error',
			textClass: 'text-error-content'
		},
		{
			id: 'borrow-pool',
			label: 'Borrow pool',
			sub: 'Pool=borrow, category blocked, no quota?',
			x: 240,
			y: 646,
			w: 270,
			h: 64,
			bgClass: 'bg-warning',
			textClass: 'text-warning-content',
			east: 'block-4',
			south: 'resolve'
		},
		{
			id: 'block-4',
			label: 'Block',
			x: 570,
			y: 662,
			w: 90,
			h: 36,
			bgClass: 'bg-error',
			textClass: 'text-error-content'
		},
		{
			id: 'resolve',
			label: 'Resolve domain',
			x: 320,
			y: 758,
			w: 200,
			h: 44,
			bgClass: 'bg-success',
			textClass: 'text-success-content'
		}
	];
	const features = [
		{
			id: 'dns-over-tls-proxy',
			title: 'DNS-over-TLS Proxy',
			desc: 'bDNS sits between you and your upstream DNS resolver, intercepting DoT queries to apply access rules.'
		},
		{
			id: 'profile-based-rules',
			title: 'Profile-Based Rules',
			desc: 'Each user profile gets its own whitelists, category blocks, and time-based schedules.'
		},
		{
			id: 'social-limits',
			title: 'Social Limits',
			desc: 'Pair with a friend to share a pool of queries. Borrow from each other when one runs out, or share a common pool.'
		},
		{
			id: 'time-blocking',
			title: 'Time Blocking',
			desc: 'Set schedules per category. Block social media after 10pm, or restrict entertainment during work hours.'
		},
		{
			id: 'category-filtering',
			title: 'Category Filtering',
			desc: 'Domains are organised into categories (ads, porn, gambling, social media, etc). Block entire categories in one click.'
		},
		{
			id: 'temporary-whitelisting',
			title: 'Temporary Whitelisting',
			desc: 'Need a blocked site just for today? Grant an end-of-day exception that resets automatically.'
		}
	];

	const flowSteps = [
		{
			id: 'identified-by-sni',
			title: 'Identified by SNI',
			body: 'Users connect via DNS-over-TLS. The TLS Server Name Indication (SNI) carries a unique profile ID.'
		},
		{
			id: 'rule-engine',
			title: 'Rule Engine',
			body: 'Every query passes through a chain of rules: permanent whitelist -> temporary whitelist -> category block -> time block -> friend pool limits.'
		},
		{
			id: 'forward-or-refuse',
			title: 'Forward or Refuse',
			body: 'If all rules pass, your query is forwarded upstream and resolved normally. Otherwise, bDNS returns a REFUSED response to the client.'
		}
	];
</script>

<div class="mx-auto max-w-3xl space-y-10 px-6 py-8">
	<section class="space-y-4 text-center">
		<h1 class="text-4xl font-bold">What is bDNS?</h1>
		<p class="mx-auto max-w-xl text-lg text-base-content/60">
			A DNS-over-TLS proxy that lets you and your friends set boundaries around how you use the
			internet - together.
		</p>
	</section>

	<section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each features as f (f.id)}
			<div class="card border border-base-300 bg-base-100 shadow-sm">
				<div class="card-body gap-2 p-5">
					<h3 class="card-title text-base">{f.title}</h3>
					<p class="text-sm text-base-content/60">{f.desc}</p>
				</div>
			</div>
		{/each}
	</section>

	<section class="space-y-5">
		<h2 class="text-center text-2xl font-bold">How it works</h2>
		<div class="flex flex-col items-center gap-4 sm:flex-row sm:items-start">
			{#each flowSteps as step, i (step.title)}
				<div class="card flex-1 border border-base-300 bg-base-100 shadow-sm">
					<div class="card-body gap-2 p-5">
						<span class="badge badge-sm badge-neutral">Step {i + 1}</span>
						<h3 class="card-title text-base">{step.title}</h3>
						<p class="text-sm text-base-content/60">{step.body}</p>
					</div>
				</div>
				{#if i < flowSteps.length - 1}
					<div class="hidden self-center text-2xl text-base-content/20 sm:block">→</div>
				{/if}
			{/each}
		</div>
	</section>

	<section class="space-y-4">
		<h2 class="text-center text-2xl font-bold">Rule engine priority</h2>
		<p class="text-center text-sm text-base-content/60">
			Every DNS query is evaluated against these rules in order. The first match determines the
			result.
		</p>
		<div class="overflow-x-auto">
			<table class="table table-zebra">
				<thead>
					<tr>
						<th>Priority</th>
						<th>Rule</th>
						<th>What it does</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><span class="badge badge-sm">1</span></td>
						<td class="font-medium">Permanent Whitelist</td>
						<td class="text-sm"
							>Always allow this domain for this profile, regardless of any other rules.</td
						>
					</tr>
					<tr>
						<td><span class="badge badge-sm">2</span></td>
						<td class="font-medium">Temporary Whitelist</td>
						<td class="text-sm">Allow this domain until end of day only.</td>
					</tr>
					<tr>
						<td><span class="badge badge-sm">3</span></td>
						<td class="font-medium">Category Block</td>
						<td class="text-sm">Block all domains in a category (e.g. ads, gambling).</td>
					</tr>
					<tr>
						<td><span class="badge badge-sm">4</span></td>
						<td class="font-medium">Time Block</td>
						<td class="text-sm"
							>Block a category during a scheduled time window (e.g. social media 10pm–6am).</td
						>
					</tr>
					<tr>
						<td><span class="badge badge-sm">5</span></td>
						<td class="font-medium">Friend Pool (Shared)</td>
						<td class="text-sm"
							>Friends share a pool of ~6k daily queries. Once exhausted, both are blocked for that
							category.</td
						>
					</tr>
					<tr>
						<td><span class="badge badge-sm">6</span></td>
						<td class="font-medium">Friend Pool (Borrow)</td>
						<td class="text-sm"
							>Each friend gets ~3k queries. When one runs out, they can borrow from the other's
							quota.</td
						>
					</tr>
					<tr>
						<td><span class="badge badge-sm badge-success">✓</span></td>
						<td class="font-medium">Allow</td>
						<td class="text-sm">No rules blocked this domain — resolve it normally.</td>
					</tr>
				</tbody>
			</table>
		</div>
	</section>

	<section class="space-y-4">
		<h2 class="text-center text-2xl font-bold">Rule engine flow</h2>
		<div class="h-[820px] w-full">
			<Svelvet controls edgeStyle="step" theme="dark" translation={{ x: 60, y: 20 }}>
				{#each nodes as n (n.id)}
					<Node
						id={n.id}
						dimensions={{ width: n.w, height: n.h }}
						position={{ x: n.x, y: n.y }}
						bgColor="transparent"
						borderColor="transparent"
						locked
					>
						<div
							class="node-inner {n.bgClass} {n.textClass} box-border flex flex-col items-center justify-center gap-0.5 rounded-lg px-3"
							style="width: {n.w}px; height: {n.h}px;"
						>
							<span class="text-center text-sm leading-tight font-semibold">{n.label}</span>
							{#if n.sub}
								<span class="text-center text-xs leading-snug opacity-80">{n.sub}</span>
							{/if}
						</div>
						<div class="anchor-top"><Anchor direction="north" invisible /></div>
						<div class="anchor-left"><Anchor direction="west" invisible /></div>
						{#if n.south}
							<div class="anchor-bottom">
								<Anchor direction="south" invisible connections={[n.south]} />
							</div>
						{/if}
						{#if n.east}
							<div class="anchor-right">
								<Anchor direction="east" invisible connections={[n.east]} />
							</div>
						{/if}
					</Node>
				{/each}
			</Svelvet>
		</div>
	</section>
</div>

<style>
	:global(.svelvet-node) {
		box-shadow: none !important;
		background: transparent !important;
		border: none !important;
	}
	.anchor-bottom {
		position: absolute;
		bottom: 0;
		left: 50%;
		transform: translateX(-50%);
	}
	.anchor-top {
		position: absolute;
		top: 0;
		left: 50%;
		transform: translateX(-50%);
	}
	.anchor-right {
		position: absolute;
		right: 0;
		top: 50%;
		transform: translateY(-50%);
	}
	.anchor-left {
		position: absolute;
		left: 0;
		top: 50%;
		transform: translateY(-50%);
	}
</style>
