import forms from '@tailwindcss/forms';
import typography from '@tailwindcss/typography';

const daisyui = require('daisyui');

export default {
	content: ['./src/**/*.{html,js,ts,svelte}'],
	plugins: [forms({ strategy: 'class' }), typography, daisyui],
	daisyui: {
		themes: [
			'light',
			{
				business: {
					/* optional overrides */
				},
			},
		],
	},
};
