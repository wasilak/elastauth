// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import starlightOpenAPI, { openAPISidebarGroups } from 'starlight-openapi';

// https://astro.build/config
export default defineConfig({
	site: process.env.ASTRO_SITE || 'https://wasilak.github.io',
	base: process.env.ASTRO_BASE || '/elastauth',
	integrations: [
		mermaid({
			theme: 'default',
			autoTheme: true,
		}),
		starlight({
			title: 'elastauth Documentation',
			description: 'A stateless authentication proxy for Elasticsearch and Kibana with pluggable authentication providers',
			logo: {
				src: './src/assets/logo.svg',
				replacesTitle: true,
			},
			social: [
				{ 
					icon: 'github', 
					label: 'GitHub', 
					href: 'https://github.com/wasilak/elastauth' 
				},
			],
			editLink: {
				baseUrl: 'https://github.com/wasilak/elastauth/edit/main/docs/',
			},
			lastUpdated: true,
			plugins: [
				// Generate the OpenAPI documentation pages
				starlightOpenAPI([
					{
						base: 'api',
						schema: './src/schemas/openapi.yaml',
						sidebar: { 
							label: 'API Reference',
							collapsed: false
						},
					},
				]),
			],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'index' },
						{ label: 'Concepts', slug: 'getting-started/concepts' },
					],
				},
				{
					label: 'Authentication Providers',
					autogenerate: { directory: 'providers' },
				},
				{
					label: 'Cache Providers',
					autogenerate: { directory: 'cache' },
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Troubleshooting', slug: 'guides/troubleshooting' },
						{ label: 'Upgrading', slug: 'guides/upgrading' },
					],
				},
				// Add the generated OpenAPI sidebar groups
				...openAPISidebarGroups,
			],
		}),
	],
});
