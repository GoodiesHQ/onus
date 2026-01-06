<script lang="ts">
	type LabelProps = {
		target: string;
		text: string;
	};

	let { target, text }: LabelProps = $props();
</script>

<label class="label block py-1" for={target}>
	<span class="label-text text-xs opacity-70">{text}</span>
</label>
