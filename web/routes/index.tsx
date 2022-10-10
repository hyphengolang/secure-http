import { Handlers, PageProps } from "$fresh/server.ts";
import Counter from "../islands/Counter.tsx";

export const handler: Handlers = {
    async GET(_, ctx) {
        try {
            const res = await fetch(`http://localhost:8080/count`);
            const { value } = await res.json();
            return ctx.render(value);
        } catch (error) {
            return ctx.render(undefined);
        }
    }
};

export default function Home({ data: value }: PageProps<number | undefined>) {
    return (
        <div class="p-4 mx-auto max-w-screen-md">
            <img
                src="/logo.svg"
                class="w-32 h-32"
                alt="the fresh logo: a sliced lemon dripping with juice"
            />
            <p class="my-6">
                Welcome to `fresh`. Try updating this message in the ./routes/index.tsx
                file, and refresh.
            </p>
            <Counter start={value ?? 0} />
        </div>
    );
}
