<script>

 function readableBytes(bytes) {
     var i = Math.floor(Math.log(bytes) / Math.log(1024)),
         sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

     return (bytes / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + sizes[i];
 }

 let files;
 let error;
 let success;
 async function handleSubmit(event) {
     error = null;
     console.log(event);
     console.log(event.target);
     let fd = new FormData(event.target);
     let resp = await fetch("__backendHost__/upload",{
         body: fd,
         method: "post",
     });

     let body = await resp.json();
     if (!resp.ok) {
         error = body.error;
         return
     }
     console.log(body);
     success = body;
 }
</script>

<div class="max-w-md rounded overflow-hidden shadow mt-8">
    <div class="px-6 py-4">
        <div class="font-bold text-xl mb-2">Welcome to super</div>
        <p class="text-gray-700 text-base">
            Please upload the files for your submission
        </p>
    </div>
    <div class="px-6 py-4">
        {#if error}
            <div class="bg-orange-100 border-l-4 border-orange-500 text-orange-700 p-4" role="alert">
                <p class="font-bold">Error</p>
                <p>{ error }</p>
            </div>
        {/if}
        {#if success}
            <div class="bg-teal-100 border-l-4 border-teal-500 text-teal-900 p-4" role="alert">
                <p class="font-bold">Success</p>
                <p>Your { success.count } file(s) have been submitted on <i>{ new Date(success.time) }</i>.</p>
                <p>Your verification ID is <code>{ success.id }</code>.</p>
            </div>
        {/if}
        <form class="w-full max-w-sm" on:submit|preventDefault="{handleSubmit}">
            <div class="flex items-center border-b border-teal-500 py-2">
                <input name="name" class="appearance-none bg-transparent border-none w-full text-gray-700 mr-3 py-1 px-2 leading-tight focus:outline-none" type="text" placeholder="Submission Name" aria-label="Submission Name">
            </div>
            <label class="my-4 w-full flex flex-col items-center px-4 py-6 bg-white text-blue rounded-lg shadow-lg tracking-wide uppercase border border-blue cursor-pointer hover:bg-teal-500 hover:text-white">
                <svg class="w-8 h-8" fill="currentColor" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
                    <path d="M16.88 9.1A4 4 0 0 1 16 17H5a5 5 0 0 1-1-9.9V7a3 3 0 0 1 4.52-2.59A4.98 4.98 0 0 1 17 8c0 .38-.04.74-.12 1.1zM11 11h3l-4-4-4 4h3v3h2v-3z" />
                </svg>
                <span class="mt-2 text-base leading-normal">Select files</span>
                <input type='file' class="hidden" name="upload[]" bind:files multiple />
            </label>
            <div class="w-full my-2">
                {#if files}
                    {#each files as file}
                        <div class="text-sm my-1">
                            <p class="text-gray-900 leading-none">{ file.name }</p>
                            <p class="text-gray-600">{ readableBytes(file.size) }</p>
                        </div>
                    {/each}
                {/if}
            </div>
            <input type="submit" class="w-full bg-teal-500 hover:bg-teal-700 text-white font-bold py-2 px-4 rounded" text="Submit">
        </form>
    </div>
</div>
 
