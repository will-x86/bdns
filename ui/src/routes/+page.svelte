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
</script>

<div class="h-full p-0">
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

				<div class="anchor-top">
					<Anchor direction="north" invisible />
				</div>
				<div class="anchor-left">
					<Anchor direction="west" invisible />
				</div>
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
