<!-- 
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. 
-->

<template>
  <!-- This component displays a 404 page not found view -->
  <v-row no-gutters class="overflow-hidden bg-containerBg" style="min-height: 100vh">
    <v-col class="d-flex align-center justify-center">
      <div class="text-center">
        <h1 class="text-h1 mt-16">Page Not Found</h1>
        <p class="text-h6 text-lightText">The page you are looking was moved, removed, <br />renamed, or might have never existed!</p>
        <v-btn variant="flat" color="primary" class="mt-2" to="/"> Back To Home </v-btn>
      </div>
    </v-col>
  </v-row>
</template>
<script setup lang="ts">
// No logic is required for this simple view component
</script>
<style lang="scss">
// These styles seem unrelated to the 404 view.
// Consider moving them to a global stylesheet or a component where they are used.
.CardMediaWrapper {
  max-width: 720px;
  margin: 0 auto;
  position: relative;
  > img {
    @media (min-width: 0px) {
      width: 250px;
      height: 130px;
    }
    @media (min-width: 768px) {
      width: 590px;
      height: 300px;
    }
  }
}
.CardMediaBuild {
  position: absolute;
  top: 60px;
  @media (min-width: 0px) {
    width: 130px;
    height: 115px;
    right: -14%;
  }
  @media (min-width: 768px) {
    width: 390px;
    height: 330px;
    right: -60%;
  }
}
</style>
