// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';

// https://astro.build/config
export default defineConfig({
	site: 'https://wasilak.github.io',
	base: '/elastauth',
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
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'index' },
					],
				},
			],
		}),
	],
});
