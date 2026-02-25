<template>
  <DeploySetup :deploy-mode="deployMode" />
</template>

<script setup>
import DeploySetup from '@/views/DeploySetup.vue'

defineProps({
  deployMode: { type: String, default: 'docker' }
})
</script>
