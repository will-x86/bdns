<div class="mx-auto max-w-2xl space-y-8 px-6 py-8">
	<section class="space-y-2 text-center">
		<h1 class="text-4xl font-bold">FAQ</h1>
		<p class="text-base-content/60">Common questions about bDNS</p>
	</section>

	<div class="join join-vertical w-full bg-base-100">
		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" checked />
			<div class="collapse-title font-semibold">What exactly is bDNS?</div>
			<div class="collapse-content space-y-2 text-sm">
				<p>
					bDNS (Bad DNS) is a <strong>DNS-over-TLS proxy</strong> that enforces access rules on your DNS
					queries. Rather than replacing your DNS server, it sits between your device and an upstream
					resolver (Cloudflare), inspecting each request and deciding whether to forward or block it.
				</p>
				<p>
					The idea is social: you pair with friend(s), and together you manage limits on things like
					social media, helping each other stay intentional about screen time.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">Why DNS-over-TLS instead of DNS-over-HTTPS?</div>
			<div class="collapse-content text-sm">
				<p>
					DNS-over-TLS (DoT) sends the <strong>TLS Server Name Indication (SNI)</strong> in the very first
					packet of the handshake. bDNS reads the SNI to identify which user profile is making the request,
					before any DNS data is exchanged. This means to use it, just configure your DNS client to use
					a unique SNI (XX.domain.com), and you're identified automatically.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">How do I connect my device?</div>
			<div class="collapse-content space-y-2 text-sm">
				<p>Set your device's Private DNS (DoT) to use your bDNS profile hostname.</p>
				<p>
					<strong>Android:</strong> Settings -> Network &amp; Internet -> Private DNS -> enter
					<code class="rounded bg-base-200 px-1 text-xs">your-profile-id.dns.example.com</code>
				</p>
				<p>
					<strong>iOS:</strong> Requires a configuration profile. Use a tool like <em>DNSecure</em>
					or install a <code>.mobileconfig</code> profile.
				</p>
				<p>
					<strong>CLI (kdig):</strong>
					<code class="rounded bg-base-200 px-1 text-xs"
						>kdig @server-ip -p 853 +tls-sni=your-profile-id.dns.example.com google.com</code
					>
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">What is my "profile ID"?</div>
			<div class="collapse-content text-sm">
				<p>
					Your profile ID is the unique ID you use as the TLS SNI when connecting. It's what bDNS
					reads from your DoT handshake to look up your rules, whitelists, and pool limits. You
					receive it when you register / in your profile. Keep it private, anyone with your profile
					ID can make queries against your limits.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">What categories can I block?</div>
			<div class="collapse-content text-sm">
				<p>
					bDNS uses the <a
						href="https://github.com/StevenBlack/hosts"
						class="link"
						target="_blank"
						rel="noopener">StevenBlack hosts blocklist</a
					>
					and other public lists to categorise domains. Categories include:
				</p>
				<ul class="mt-1 ml-5 list-disc space-y-1">
					<li>Advertising &amp; trackers</li>
					<li>Pornography</li>
					<li>Gambling</li>
					<li>Social media</li>
					<li>Malware &amp; phishing</li>
					<li>Fake news</li>
				</ul>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">
				What's the difference between Shared and Borrow pools?
			</div>
			<div class="collapse-content space-y-2 text-sm">
				<p>
					<strong>Shared pool:</strong> You and your friend share a single pool of queries (e.g.
					6,000/day). If either of you uses them all, <em>both</em> of you are blocked for that category
					until the quota resets at midnight.
				</p>
				<p>
					<strong>Borrow pool:</strong> Each of you gets your own quota (e.g. 3,000/day). When one person
					runs out, they can temporarily borrow from the other's remaining quota. Borrowed limit is returned
					at end of day.
				</p>
				<p>
					Shared encourages balance; Borrow lets the lighter user help the heavier one without
					penalty.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">How do time-based blocks work?</div>
			<div class="collapse-content text-sm">
				<p>
					You can set a schedule per category. For example, "block social media from 22:00 to
					07:00." bDNS checks the current time (adjusted to your timezone) against your schedule. If
					the current time falls within a blocked window, queries to domains in that category are
					refused. Outside the window, the query passes through to the next rule.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">What happens when a domain is blocked?</div>
			<div class="collapse-content text-sm">
				<p>
					bDNS sends a <strong>REFUSED</strong> DNS response. Your browser or app will see this as a DNS
					failure, the site simply won't load.
				</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">Which upstream DNS server does bDNS use?</div>
			<div class="collapse-content text-sm">
				<p>bDNS forwards allowed queries to Cloudflare</p>
			</div>
		</div>

		<div class="collapse-arrow collapse join-item border border-base-300">
			<input type="radio" name="faq-accordion" />
			<div class="collapse-title font-semibold">
				I'm being blocked from a site I need. What can I do?
			</div>
			<div class="collapse-content space-y-2 text-sm">
				<p>Two options:</p>
				<p>
					<strong>Permanent whitelist:</strong> Add the domain to your profile's permanent whitelist.
					It will always resolve, regardless of categories or schedules.
				</p>
				<p>
					<strong>Temporary whitelist:</strong> Add a time-limited exception that expires at end of day.
					Great for one-off access without changing your permanent rules.
				</p>
			</div>
		</div>
	</div>
</div>
