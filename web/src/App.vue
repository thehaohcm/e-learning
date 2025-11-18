<template>
  <div id="app">
    <nav>
      <router-link to="/">Dashboard</router-link> |
      <router-link to="/learn">Learn</router-link> |
      <router-link to="/quiz">Quiz</router-link>
    </nav>
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

nav {
  padding: 30px;
}

nav a {
  font-weight: bold;
  color: #2c3e50;
  text-decoration: none;
  margin: 0 10px;
}

nav a.router-link-exact-active {
  color: #42b983;
}
</style>