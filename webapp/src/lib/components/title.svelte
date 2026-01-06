<script lang="ts">
	type TitleProps = {
		text: string;
	};

	let { text }: TitleProps = $props();
</script>

<h1 class="mb-4 font-sans text-xl font-bold font-stretch-100%">{text}</h1>
