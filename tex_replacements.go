package main

var TexTemplate = `\documentclass{article}
\usepackage{pagecolor}
\begin{document}

$%s$

\pagenumbering{gobble}
\end{document}`


var TexReplacements = map[string]string{
	"Γ": " \\Gamma ",
	"Δ": " \\Delta ",
	"Θ": " \\Theta ",
	"Λ": " \\Lambda ",
	"Ξ": " \\Xi ",
	"Π": " \\Pi ",
	"Σ": " \\Sigma",
	"Υ": " \\Upsilon",
	"Φ": " \\Phi ",
	"Ψ": " \\Psi ",
	"Ω": " \\Omega",
	"α": " \\alpha ",
	"β": " \\beta ",
	"γ": " \\gamma ",
	"δ": " \\delta ",
	"ε": " \\epsilon ",
	"ζ": " \\zeta ",
	"η": " \\eta ",
	"θ": " \\theta ",
	"ι": " \\iota ",
	"κ": " \\kappa ",
	"λ": " \\lambda ",
	"μ": " \\mu ",
	"ν": " \\nu ",
	"ξ": " \\xi ",
	"π": " \\pi ",
	"ρ": " \\rho ",
	"ς": " \\varsigma ",
	"σ": " \\sigma ",
	"τ": " \\tau ",
	"υ": " \\upsilon ",
	"φ": " \\phi ",
	"χ": " \\chi ",
	"ψ": " \\psi ",
	"ω": " \\omega ",
	"×": " \\times ",
	"÷": " \\div ",
	"ש": " \\shin ",
	"א": " \\alef ",
	"ב": " \\beth ",
	"ג": " \\gimel ",
	"ד": " \\daleth ",
	"ל": " \\lamed ",
	"מ": " \\mim ",
	"ם": " \\mim ",
	"ע": " \\ayin ",
	"צ": " \\tsadi ",
	"ץ": " \\tsadi ",
	"ק": " \\qof ",
	"≠": " \\neq ",
	"·": " \\cdot ",
	"•": " \\cdot ",
}