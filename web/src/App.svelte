<script>
 import {Route,router} from 'tinro';
 import Submit from './Submit.svelte';
 import Admin from './Admin.svelte';
 router.useHashNavigation(); // enable hash navigation method
</script>

<style global>
 /* purgecss start ignore */
 @tailwind base;
 @tailwind components;
 /* purgecss end ignore */
 @tailwind utilities;
</style>
<main class="w-screen flex justify-center">
    <Route path="/">
      <Submit></Submit>
    </Route>
    <Route path="/admin">
      <Admin></Admin>
    </Route>
</main>
<p class="text-center text-gray-500 text-xs w-screen">
    <a href="https://github.com/theSuess/super">super</a> - the simple submission platform.
</p>
