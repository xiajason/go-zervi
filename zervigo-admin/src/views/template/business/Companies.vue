<template>
  <div class="companies-page">
    <DynamicTable
      table-name="company"
      model-label="企业"
      api-path="/api/v1/company/index"
      :can-create="true"
      :can-update="true"
      :can-delete="true"
    />
  </div>
</template>

<script setup lang="ts">
import DynamicTable from '@/components/DynamicTable.vue'
</script>

<style scoped>
.companies-page {
  padding: 20px;
}
</style>

