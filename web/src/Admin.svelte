<script>
 let submissions;
 let token;
 let error = "Please enter your password to continue";

 async function fetchSubmissions(event) {
     error = null;
     let pw = event.target.password.value;
     let resp = await fetch("__backendHost__/list",{
         headers: {
             "X-Auth-Token": pw
         }
     })
     let body = await resp.json();
     if (!resp.ok) {
         error = body.error;
         return;
     }
     console.log(resp);
     token = resp.headers.get("X-Download-Token");
     submissions = body;
 }
</script>

<div class="rounded overflow-hidden shadow mt-8">
    <div class="px-6 py-4">
        <div class="font-bold text-xl mb-2">Admin Interface</div>
    </div>
    {#if error}
        <div class="px-6 py-4">
            <div class="bg-orange-100 border-l-4 border-orange-500 text-orange-700 p-4" role="alert">
                <p class="font-bold">Error</p>
                <p>{ error }</p>
            </div>
        </div>
    {/if}
    {#if !submissions}
        <div class="px-6 py-4">
            <form on:submit|preventDefault="{fetchSubmissions}">
                <div class="flex items-center border-b border-teal-500 py-2">
                    <input name="password" class="appearance-none bg-transparent border-none w-full text-gray-700 mr-3 py-1 px-2 leading-tight focus:outline-none" type="password" placeholder="Password" aria-label="Password">
                </div>
                <input type="submit" class="w-full bg-teal-500 hover:bg-teal-700 text-white font-bold py-2 px-4 rounded" value="Fetch Submissions">
            </form>
        </div>
    {:else}
        <div class="px-6 py-4">
            <div class="font-bold text-xl mb-2">List of Submissions</div>
            <div class="w-full my-2">
                {#each submissions as sub}
                    <div class="w-full rounded overflow-hidden shadow my-1">
                        <div class="p-3">
                            <h3 class="mr-10 text-lg truncate-2nd">
                                { sub.name }
                            </h3>
                            <p class="text-gray-600">{ new Date(sub.time) }</p>
                            <div>
                                <ul>
                                {#each Object.entries(sub.files) as [id,name]}
                                    <li class="text-teal-500 hover:text-teal-700">
                                        <a href="__backendHost__/download/{sub.id}/{id}?token={token}&name={name}" tinro-ignore>{ name }</a>
                                    </li>
                                {/each}
                                </ul>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>
