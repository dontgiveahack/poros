import { useState } from "react"

function App() {
	const [count, setCount] = useState(0)

	return (
		<main style={{ fontFamily: "system-ui, sans-serif", padding: "2rem"}}>
			<h1>Póros Manager</h1>
			<p>Personal finances, plaintext as source of truth.</p>
			<button onClick={() => setCount((c) => c + 1)}>
				CLI clicks: {count}
			</button>
		</main>
	)
}

export default App
